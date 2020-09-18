// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by client-gen. DO NOT EDIT.

package versioned

import (
	"fmt"

	crdv1alpha1 "go.pinniped.dev/generated/1.17/client/clientset/versioned/typed/crdpinniped/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.17/client/clientset/versioned/typed/idp/v1alpha1"
	loginv1alpha1 "go.pinniped.dev/generated/1.17/client/clientset/versioned/typed/login/v1alpha1"
	pinnipedv1alpha1 "go.pinniped.dev/generated/1.17/client/clientset/versioned/typed/pinniped/v1alpha1"
	discovery "k8s.io/client-go/discovery"
	rest "k8s.io/client-go/rest"
	flowcontrol "k8s.io/client-go/util/flowcontrol"
)

type Interface interface {
	Discovery() discovery.DiscoveryInterface
	CrdV1alpha1() crdv1alpha1.CrdV1alpha1Interface
	IDPV1alpha1() idpv1alpha1.IDPV1alpha1Interface
	LoginV1alpha1() loginv1alpha1.LoginV1alpha1Interface
	PinnipedV1alpha1() pinnipedv1alpha1.PinnipedV1alpha1Interface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	*discovery.DiscoveryClient
	crdV1alpha1      *crdv1alpha1.CrdV1alpha1Client
	iDPV1alpha1      *idpv1alpha1.IDPV1alpha1Client
	loginV1alpha1    *loginv1alpha1.LoginV1alpha1Client
	pinnipedV1alpha1 *pinnipedv1alpha1.PinnipedV1alpha1Client
}

// CrdV1alpha1 retrieves the CrdV1alpha1Client
func (c *Clientset) CrdV1alpha1() crdv1alpha1.CrdV1alpha1Interface {
	return c.crdV1alpha1
}

// IDPV1alpha1 retrieves the IDPV1alpha1Client
func (c *Clientset) IDPV1alpha1() idpv1alpha1.IDPV1alpha1Interface {
	return c.iDPV1alpha1
}

// LoginV1alpha1 retrieves the LoginV1alpha1Client
func (c *Clientset) LoginV1alpha1() loginv1alpha1.LoginV1alpha1Interface {
	return c.loginV1alpha1
}

// PinnipedV1alpha1 retrieves the PinnipedV1alpha1Client
func (c *Clientset) PinnipedV1alpha1() pinnipedv1alpha1.PinnipedV1alpha1Interface {
	return c.pinnipedV1alpha1
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
			return nil, fmt.Errorf("Burst is required to be greater than 0 when RateLimiter is not set and QPS is set to greater than 0")
		}
		configShallowCopy.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(configShallowCopy.QPS, configShallowCopy.Burst)
	}
	var cs Clientset
	var err error
	cs.crdV1alpha1, err = crdv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.iDPV1alpha1, err = idpv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.loginV1alpha1, err = loginv1alpha1.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}
	cs.pinnipedV1alpha1, err = pinnipedv1alpha1.NewForConfig(&configShallowCopy)
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
	cs.crdV1alpha1 = crdv1alpha1.NewForConfigOrDie(c)
	cs.iDPV1alpha1 = idpv1alpha1.NewForConfigOrDie(c)
	cs.loginV1alpha1 = loginv1alpha1.NewForConfigOrDie(c)
	cs.pinnipedV1alpha1 = pinnipedv1alpha1.NewForConfigOrDie(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClientForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.crdV1alpha1 = crdv1alpha1.New(c)
	cs.iDPV1alpha1 = idpv1alpha1.New(c)
	cs.loginV1alpha1 = loginv1alpha1.New(c)
	cs.pinnipedV1alpha1 = pinnipedv1alpha1.New(c)

	cs.DiscoveryClient = discovery.NewDiscoveryClient(c)
	return &cs
}
