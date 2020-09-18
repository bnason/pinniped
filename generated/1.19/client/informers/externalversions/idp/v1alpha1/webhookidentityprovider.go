// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	idpv1alpha1 "go.pinniped.dev/generated/1.19/apis/idp/v1alpha1"
	versioned "go.pinniped.dev/generated/1.19/client/clientset/versioned"
	internalinterfaces "go.pinniped.dev/generated/1.19/client/informers/externalversions/internalinterfaces"
	v1alpha1 "go.pinniped.dev/generated/1.19/client/listers/idp/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// WebhookIdentityProviderInformer provides access to a shared informer and lister for
// WebhookIdentityProviders.
type WebhookIdentityProviderInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.WebhookIdentityProviderLister
}

type webhookIdentityProviderInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewWebhookIdentityProviderInformer constructs a new informer for WebhookIdentityProvider type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewWebhookIdentityProviderInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredWebhookIdentityProviderInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredWebhookIdentityProviderInformer constructs a new informer for WebhookIdentityProvider type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredWebhookIdentityProviderInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IDPV1alpha1().WebhookIdentityProviders(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.IDPV1alpha1().WebhookIdentityProviders(namespace).Watch(context.TODO(), options)
			},
		},
		&idpv1alpha1.WebhookIdentityProvider{},
		resyncPeriod,
		indexers,
	)
}

func (f *webhookIdentityProviderInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredWebhookIdentityProviderInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *webhookIdentityProviderInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&idpv1alpha1.WebhookIdentityProvider{}, f.defaultInformer)
}

func (f *webhookIdentityProviderInformer) Lister() v1alpha1.WebhookIdentityProviderLister {
	return v1alpha1.NewWebhookIdentityProviderLister(f.Informer().GetIndexer())
}
