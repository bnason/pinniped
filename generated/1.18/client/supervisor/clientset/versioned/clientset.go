// Copyright 2020-2022 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"

	configv1alpha1 "go.pinniped.dev/generated/1.18/client/supervisor/clientset/versioned/typed/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.18/client/supervisor/clientset/versioned/typed/idp/v1alpha1"
	oauthv1alpha1 "go.pinniped.dev/generated/1.18/client/supervisor/clientset/versioned/typed/oauth/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface
	IDPV1alpha1() idpv1alpha1.IDPV1alpha1Interface
	OauthV1alpha1() oauthv1alpha1.OauthV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	configV1alpha1 *configv1alpha1.ConfigV1alpha1Client
	iDPV1alpha1    *idpv1alpha1.IDPV1alpha1Client
	oauthV1alpha1  *oauthv1alpha1.OauthV1alpha1Client
}

// ConfigV1alpha1 retrieves the ConfigV1alpha1Client
func (c *Clientset) ConfigV1alpha1() configv1alpha1.ConfigV1alpha1Interface {
	return c.configV1alpha1
}

// IDPV1alpha1 retrieves the IDPV1alpha1Client
func (c *Clientset) IDPV1alpha1() idpv1alpha1.IDPV1alpha1Interface {
	return c.iDPV1alpha1
}

// OauthV1alpha1 retrieves the OauthV1alpha1Client
func (c *Clientset) OauthV1alpha1() oauthv1alpha1.OauthV1alpha1Interface {
	return c.oauthV1alpha1
}

// Discovery retrieves the DiscoveryClient
func (c *Clientset) Discovery() discovery.DiscoveryInterface {
	if c == nil {
		return nil
	}
	return c.DiscoveryClient
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c
	if configShallowCopy.RateLimiter == nil && configShallowCopy.QPS > 0 {
		if configShallowCopy.Burst <= 0 {
			return nil, fmt.Errorf("burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.configV1alpha1, err = configv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.iDPV1alpha1, err = idpv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.oauthV1alpha1, err = oauthv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	cs.DiscoveryClient, err = discovery.NewDiscoveryClientForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.configV1alpha1 = configv1alpha1.NewForConfigOrDie(c)
	cs.iDPV1alpha1 = idpv1alpha1.NewForConfigOrDie(c)
	cs.oauthV1alpha1 = oauthv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.configV1alpha1 = configv1alpha1.New(c)
	cs.iDPV1alpha1 = idpv1alpha1.New(c)
	cs.oauthV1alpha1 = oauthv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
