// Copyright 2020 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package externalversions

import (
	"fmt"

	loginv1alpha1 "go.pinniped.dev/generated/1.18/apis/concierge/login/v1alpha1"
	v1alpha1 "go.pinniped.dev/generated/1.18/apis/config/v1alpha1"
	idpv1alpha1 "go.pinniped.dev/generated/1.18/apis/idp/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	cache "k8s.io/client-go/tools/cache"
)

// GenericInformer is type of SharedIndexInformer which will locate and delegate to other
// sharedInformers based on type
type GenericInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() cache.GenericLister
}

type genericInformer struct {
	informer cache.SharedIndexInformer
	resource schema.GroupResource
}

// Informer returns the SharedIndexInformer.
func (f *genericInformer) Informer() cache.SharedIndexInformer {
	return f.informer
}

// Lister returns the GenericLister.
func (f *genericInformer) Lister() cache.GenericLister {
	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
}

// ForResource gives generic access to a shared informer of the matching type
// TODO extend this to unknown resources with a client pool
func (f *sharedInformerFactory) ForResource(resource schema.GroupVersionResource) (GenericInformer, error) {
	switch resource {
	// Group=config.pinniped.dev, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithResource("credentialissuerconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Config().V1alpha1().CredentialIssuerConfigs().Informer()}, nil
	case v1alpha1.SchemeGroupVersion.WithResource("oidcproviderconfigs"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Config().V1alpha1().OIDCProviderConfigs().Informer()}, nil

		// Group=idp.pinniped.dev, Version=v1alpha1
	case idpv1alpha1.SchemeGroupVersion.WithResource("webhookidentityproviders"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.IDP().V1alpha1().WebhookIdentityProviders().Informer()}, nil

		// Group=login.concierge.pinniped.dev, Version=v1alpha1
	case loginv1alpha1.SchemeGroupVersion.WithResource("tokencredentialrequests"):
		return &genericInformer{resource: resource.GroupResource(), informer: f.Login().V1alpha1().TokenCredentialRequests().Informer()}, nil

	}

	return nil, fmt.Errorf("no informer found for %v", resource)
}
