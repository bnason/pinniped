// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package auth provides a handler for the OIDC authorization endpoint.
package auth

import (
	"fmt"
	"net/http"
	"time"

	coreosoidc "github.com/coreos/go-oidc/v3/oidc"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"k8s.io/apiserver/pkg/authentication/authenticator"

	supervisoroidc "go.pinniped.dev/generated/latest/apis/supervisor/oidc"
	"go.pinniped.dev/internal/httputil/httperr"
	"go.pinniped.dev/internal/httputil/securityheader"
	"go.pinniped.dev/internal/oidc"
	"go.pinniped.dev/internal/oidc/csrftoken"
	"go.pinniped.dev/internal/oidc/downstreamsession"
	"go.pinniped.dev/internal/oidc/provider"
	"go.pinniped.dev/internal/plog"
	"go.pinniped.dev/pkg/oidcclient/nonce"
	"go.pinniped.dev/pkg/oidcclient/pkce"
)

func NewHandler(
	downstreamIssuer string,
	idpLister oidc.UpstreamIdentityProvidersLister,
	oauthHelperWithoutStorage fosite.OAuth2Provider,
	oauthHelperWithStorage fosite.OAuth2Provider,
	generateCSRF func() (csrftoken.CSRFToken, error),
	generatePKCE func() (pkce.Code, error),
	generateNonce func() (nonce.Nonce, error),
	upstreamStateEncoder oidc.Encoder,
	cookieCodec oidc.Codec,
) http.Handler {
	return securityheader.Wrap(httperr.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			// https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest
			// Authorization Servers MUST support the use of the HTTP GET and POST methods defined in
			// RFC 2616 [RFC2616] at the Authorization Endpoint.
			return httperr.Newf(http.StatusMethodNotAllowed, "%s (try GET or POST)", r.Method)
		}

		oidcUpstream, ldapUpstream, err := chooseUpstreamIDP(idpLister)
		if err != nil {
			plog.WarningErr("authorize upstream config", err)
			return err
		}

		if oidcUpstream != nil {
			if len(r.Header.Values(supervisoroidc.AuthorizeUsernameHeaderName)) > 0 {
				// The client set a username header, so they are trying to log in with a username/password.
				return handleAuthRequestForOIDCUpstreamPasswordGrant(r, w, oauthHelperWithStorage, oidcUpstream)
			}
			return handleAuthRequestForOIDCUpstreamAuthcodeGrant(r, w,
				oauthHelperWithoutStorage,
				generateCSRF, generateNonce, generatePKCE,
				oidcUpstream,
				downstreamIssuer,
				upstreamStateEncoder,
				cookieCodec,
			)
		}
		return handleAuthRequestForLDAPUpstream(r, w,
			oauthHelperWithStorage,
			ldapUpstream,
		)
	}))
}

func handleAuthRequestForLDAPUpstream(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	ldapUpstream provider.UpstreamLDAPIdentityProviderI,
) error {
	authorizeRequester, created := newAuthorizeRequest(r, w, oauthHelper)
	if !created {
		return nil
	}

	username, password, hadUsernamePasswordValues := requireNonEmptyUsernameAndPasswordHeaders(r, w, oauthHelper, authorizeRequester)
	if !hadUsernamePasswordValues {
		return nil
	}

	authenticateResponse, authenticated, err := ldapUpstream.AuthenticateUser(r.Context(), username, password)
	if err != nil {
		plog.WarningErr("unexpected error during upstream LDAP authentication", err, "upstreamName", ldapUpstream.GetName())
		return httperr.New(http.StatusBadGateway, "unexpected error during upstream authentication")
	}
	if !authenticated {
		plog.Debug("failed upstream LDAP authentication", "upstreamName", ldapUpstream.GetName())
		// Return an error according to OIDC spec 3.1.2.6 (second paragraph).
		err = errors.WithStack(fosite.ErrAccessDenied.WithHintf("Username/password not accepted by LDAP provider."))
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil
	}

	subject := downstreamSubjectFromUpstreamLDAP(ldapUpstream, authenticateResponse)
	username = authenticateResponse.User.GetName()
	groups := authenticateResponse.User.GetGroups()

	return makeDownstreamSessionAndReturnAuthcodeRedirect(r, w, oauthHelper, authorizeRequester, subject, username, groups)
}

func handleAuthRequestForOIDCUpstreamPasswordGrant(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	oidcUpstream provider.UpstreamOIDCIdentityProviderI,
) error {
	authorizeRequester, created := newAuthorizeRequest(r, w, oauthHelper)
	if !created {
		return nil
	}

	username, password, hadUsernamePasswordValues := requireNonEmptyUsernameAndPasswordHeaders(r, w, oauthHelper, authorizeRequester)
	if !hadUsernamePasswordValues {
		return nil
	}

	if !oidcUpstream.AllowsPasswordGrant() {
		// Return a user-friendly error for this case which is entirely within our control.
		err := errors.WithStack(fosite.ErrAccessDenied.
			WithHint("Resource owner password credentials grant is not allowed for this upstream provider according to its configuration."),
		)
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil
	}

	token, err := oidcUpstream.PasswordCredentialsGrantAndValidateTokens(r.Context(), username, password)
	if err != nil {
		// Upstream password grant errors can be generic errors (e.g. a network failure) or can be oauth2.RetrieveError errors
		// which represent the http response from the upstream server. These could be a 5XX or some other unexpected error,
		// or could be a 400 with a JSON body as described by https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
		// which notes that wrong resource owner credentials should result in an "invalid_grant" error.
		// However, the exact response is undefined in the sense that there is no such thing as a password grant in
		// the OIDC spec, so we don't try too hard to read the upstream errors in this case. (E.g. Dex departs from the
		// spec and returns something other than an "invalid_grant" error for bad resource owner credentials.)
		// Return an error according to OIDC spec 3.1.2.6 (second paragraph).
		err := errors.WithStack(fosite.ErrAccessDenied.WithDebug(err.Error())) // WithDebug hides the error from the client
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil
	}

	subject, username, groups, err := downstreamsession.GetDownstreamIdentityFromUpstreamIDToken(oidcUpstream, token.IDToken.Claims)
	if err != nil {
		return err
	}

	return makeDownstreamSessionAndReturnAuthcodeRedirect(r, w, oauthHelper, authorizeRequester, subject, username, groups)
}

func handleAuthRequestForOIDCUpstreamAuthcodeGrant(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	generateCSRF func() (csrftoken.CSRFToken, error),
	generateNonce func() (nonce.Nonce, error),
	generatePKCE func() (pkce.Code, error),
	oidcUpstream provider.UpstreamOIDCIdentityProviderI,
	downstreamIssuer string,
	upstreamStateEncoder oidc.Encoder,
	cookieCodec oidc.Codec,
) error {
	authorizeRequester, created := newAuthorizeRequest(r, w, oauthHelper)
	if !created {
		return nil
	}

	now := time.Now()
	_, err := oauthHelper.NewAuthorizeResponse(r.Context(), authorizeRequester, &openid.DefaultSession{
		Claims: &jwt.IDTokenClaims{
			// Temporary claim values to allow `NewAuthorizeResponse` to perform other OIDC validations.
			Subject:     "none",
			AuthTime:    now,
			RequestedAt: now,
		},
	})
	if err != nil {
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil
	}

	csrfValue, nonceValue, pkceValue, err := generateValues(generateCSRF, generateNonce, generatePKCE)
	if err != nil {
		plog.Error("authorize generate error", err)
		return err
	}
	csrfFromCookie := readCSRFCookie(r, cookieCodec)
	if csrfFromCookie != "" {
		csrfValue = csrfFromCookie
	}

	upstreamOAuthConfig := oauth2.Config{
		ClientID: oidcUpstream.GetClientID(),
		Endpoint: oauth2.Endpoint{
			AuthURL: oidcUpstream.GetAuthorizationURL().String(),
		},
		RedirectURL: fmt.Sprintf("%s/callback", downstreamIssuer),
		Scopes:      oidcUpstream.GetScopes(),
	}

	encodedStateParamValue, err := upstreamStateParam(
		authorizeRequester,
		oidcUpstream.GetName(),
		nonceValue,
		csrfValue,
		pkceValue,
		upstreamStateEncoder,
	)
	if err != nil {
		plog.Error("authorize upstream state param error", err)
		return err
	}

	if csrfFromCookie == "" {
		// We did not receive an incoming CSRF cookie, so write a new one.
		err := addCSRFSetCookieHeader(w, csrfValue, cookieCodec)
		if err != nil {
			plog.Error("error setting CSRF cookie", err)
			return err
		}
	}

	authCodeOptions := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		nonceValue.Param(),
		pkceValue.Challenge(),
		pkceValue.Method(),
	}

	promptParam := r.Form.Get("prompt")
	if promptParam != "" && oidc.ScopeWasRequested(authorizeRequester, coreosoidc.ScopeOpenID) {
		authCodeOptions = append(authCodeOptions, oauth2.SetAuthURLParam("prompt", promptParam))
	}

	http.Redirect(w, r,
		upstreamOAuthConfig.AuthCodeURL(
			encodedStateParamValue,
			authCodeOptions...,
		),
		302,
	)

	return nil
}

func makeDownstreamSessionAndReturnAuthcodeRedirect(
	r *http.Request,
	w http.ResponseWriter,
	oauthHelper fosite.OAuth2Provider,
	authorizeRequester fosite.AuthorizeRequester,
	subject string,
	username string,
	groups []string,
) error {
	openIDSession := downstreamsession.MakeDownstreamSession(subject, username, groups)

	authorizeResponder, err := oauthHelper.NewAuthorizeResponse(r.Context(), authorizeRequester, openIDSession)
	if err != nil {
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil
	}

	oauthHelper.WriteAuthorizeResponse(w, authorizeRequester, authorizeResponder)

	return nil
}

func requireNonEmptyUsernameAndPasswordHeaders(r *http.Request, w http.ResponseWriter, oauthHelper fosite.OAuth2Provider, authorizeRequester fosite.AuthorizeRequester) (string, string, bool) {
	username := r.Header.Get(supervisoroidc.AuthorizeUsernameHeaderName)
	password := r.Header.Get(supervisoroidc.AuthorizePasswordHeaderName)
	if username == "" || password == "" {
		// Return an error according to OIDC spec 3.1.2.6 (second paragraph).
		err := errors.WithStack(fosite.ErrAccessDenied.WithHintf("Missing or blank username or password."))
		plog.Info("authorize response error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return "", "", false
	}
	return username, password, true
}

func newAuthorizeRequest(r *http.Request, w http.ResponseWriter, oauthHelper fosite.OAuth2Provider) (fosite.AuthorizeRequester, bool) {
	authorizeRequester, err := oauthHelper.NewAuthorizeRequest(r.Context(), r)
	if err != nil {
		plog.Info("authorize request error", oidc.FositeErrorForLog(err)...)
		oauthHelper.WriteAuthorizeError(w, authorizeRequester, err)
		return nil, false
	}

	// Automatically grant the openid, offline_access, and pinniped:request-audience scopes, but only if they were requested.
	// Grant the openid scope (for now) if they asked for it so that `NewAuthorizeResponse` will perform its OIDC validations.
	// There don't seem to be any validations inside `NewAuthorizeResponse` related to the offline_access scope
	// at this time, however we will temporarily grant the scope just in case that changes in a future release of fosite.
	downstreamsession.GrantScopesIfRequested(authorizeRequester)

	return authorizeRequester, true
}

func readCSRFCookie(r *http.Request, codec oidc.Decoder) csrftoken.CSRFToken {
	receivedCSRFCookie, err := r.Cookie(oidc.CSRFCookieName)
	if err != nil {
		// Error means that the cookie was not found
		return ""
	}

	var csrfFromCookie csrftoken.CSRFToken
	err = codec.Decode(oidc.CSRFCookieEncodingName, receivedCSRFCookie.Value, &csrfFromCookie)
	if err != nil {
		// We can ignore any errors and just make a new cookie. Hopefully this will
		// make the user experience better if, for example, the server rotated
		// cookie signing keys and then a user submitted a very old cookie.
		return ""
	}

	return csrfFromCookie
}

// Select either an OIDC or an LDAP IDP, or return an error.
func chooseUpstreamIDP(idpLister oidc.UpstreamIdentityProvidersLister) (provider.UpstreamOIDCIdentityProviderI, provider.UpstreamLDAPIdentityProviderI, error) {
	oidcUpstreams := idpLister.GetOIDCIdentityProviders()
	ldapUpstreams := idpLister.GetLDAPIdentityProviders()
	switch {
	case len(oidcUpstreams)+len(ldapUpstreams) == 0:
		return nil, nil, httperr.New(
			http.StatusUnprocessableEntity,
			"No upstream providers are configured",
		)
	case len(oidcUpstreams)+len(ldapUpstreams) > 1:
		var upstreamIDPNames []string
		for _, idp := range oidcUpstreams {
			upstreamIDPNames = append(upstreamIDPNames, idp.GetName())
		}
		for _, idp := range ldapUpstreams {
			upstreamIDPNames = append(upstreamIDPNames, idp.GetName())
		}
		plog.Warning("Too many upstream providers are configured (found: %s)", upstreamIDPNames)
		return nil, nil, httperr.New(
			http.StatusUnprocessableEntity,
			"Too many upstream providers are configured (support for multiple upstreams is not yet implemented)",
		)
	case len(oidcUpstreams) == 1:
		return oidcUpstreams[0], nil, nil
	default:
		return nil, ldapUpstreams[0], nil
	}
}

func generateValues(
	generateCSRF func() (csrftoken.CSRFToken, error),
	generateNonce func() (nonce.Nonce, error),
	generatePKCE func() (pkce.Code, error),
) (csrftoken.CSRFToken, nonce.Nonce, pkce.Code, error) {
	csrfValue, err := generateCSRF()
	if err != nil {
		return "", "", "", httperr.Wrap(http.StatusInternalServerError, "error generating CSRF token", err)
	}
	nonceValue, err := generateNonce()
	if err != nil {
		return "", "", "", httperr.Wrap(http.StatusInternalServerError, "error generating nonce param", err)
	}
	pkceValue, err := generatePKCE()
	if err != nil {
		return "", "", "", httperr.Wrap(http.StatusInternalServerError, "error generating PKCE param", err)
	}
	return csrfValue, nonceValue, pkceValue, nil
}

func upstreamStateParam(
	authorizeRequester fosite.AuthorizeRequester,
	upstreamName string,
	nonceValue nonce.Nonce,
	csrfValue csrftoken.CSRFToken,
	pkceValue pkce.Code,
	encoder oidc.Encoder,
) (string, error) {
	stateParamData := oidc.UpstreamStateParamData{
		AuthParams:    authorizeRequester.GetRequestForm().Encode(),
		UpstreamName:  upstreamName,
		Nonce:         nonceValue,
		CSRFToken:     csrfValue,
		PKCECode:      pkceValue,
		FormatVersion: oidc.UpstreamStateParamFormatVersion,
	}
	encodedStateParamValue, err := encoder.Encode(oidc.UpstreamStateParamEncodingName, stateParamData)
	if err != nil {
		return "", httperr.Wrap(http.StatusInternalServerError, "error encoding upstream state param", err)
	}
	return encodedStateParamValue, nil
}

func addCSRFSetCookieHeader(w http.ResponseWriter, csrfValue csrftoken.CSRFToken, codec oidc.Encoder) error {
	encodedCSRFValue, err := codec.Encode(oidc.CSRFCookieEncodingName, csrfValue)
	if err != nil {
		return httperr.Wrap(http.StatusInternalServerError, "error encoding CSRF cookie", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oidc.CSRFCookieName,
		Value:    encodedCSRFValue,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
		Path:     "/",
	})

	return nil
}

func downstreamSubjectFromUpstreamLDAP(ldapUpstream provider.UpstreamLDAPIdentityProviderI, authenticateResponse *authenticator.Response) string {
	ldapURL := *ldapUpstream.GetURL()
	q := ldapURL.Query()
	q.Set(oidc.IDTokenSubjectClaim, authenticateResponse.User.GetUID())
	ldapURL.RawQuery = q.Encode()
	return ldapURL.String()
}
