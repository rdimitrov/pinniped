// Copyright 2020-2021 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	internalinterfaces "go.pinniped.dev/generated/1.17/client/supervisor/informers/externalversions/internalinterfaces"
)

// Interface provides access to all the informers in this group version.
type Interface interface {
	// ActiveDirectoryIdentityProviders returns a ActiveDirectoryIdentityProviderInformer.
	ActiveDirectoryIdentityProviders() ActiveDirectoryIdentityProviderInformer
	// LDAPIdentityProviders returns a LDAPIdentityProviderInformer.
	LDAPIdentityProviders() LDAPIdentityProviderInformer
	// OIDCIdentityProviders returns a OIDCIdentityProviderInformer.
	OIDCIdentityProviders() OIDCIdentityProviderInformer
}

type version struct {
	factory          internalinterfaces.SharedInformerFactory
	namespace        string
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// New returns a new Interface.
func New(f internalinterfaces.SharedInformerFactory, namespace string, tweakListOptions internalinterfaces.TweakListOptionsFunc) Interface {
	return &version{factory: f, namespace: namespace, tweakListOptions: tweakListOptions}
}

// ActiveDirectoryIdentityProviders returns a ActiveDirectoryIdentityProviderInformer.
func (v *version) ActiveDirectoryIdentityProviders() ActiveDirectoryIdentityProviderInformer {
	return &activeDirectoryIdentityProviderInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// LDAPIdentityProviders returns a LDAPIdentityProviderInformer.
func (v *version) LDAPIdentityProviders() LDAPIdentityProviderInformer {
	return &lDAPIdentityProviderInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}

// OIDCIdentityProviders returns a OIDCIdentityProviderInformer.
func (v *version) OIDCIdentityProviders() OIDCIdentityProviderInformer {
	return &oIDCIdentityProviderInformer{factory: v.factory, namespace: v.namespace, tweakListOptions: v.tweakListOptions}
}
