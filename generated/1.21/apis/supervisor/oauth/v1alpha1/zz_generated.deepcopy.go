//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Copyright 2020-2022 the Pinniped contributors. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCClient) DeepCopyInto(out *OIDCClient) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCClient.
func (in *OIDCClient) DeepCopy() *OIDCClient {
	if in == nil {
		return nil
	}
	out := new(OIDCClient)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCClient) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCClientList) DeepCopyInto(out *OIDCClientList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]OIDCClient, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCClientList.
func (in *OIDCClientList) DeepCopy() *OIDCClientList {
	if in == nil {
		return nil
	}
	out := new(OIDCClientList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *OIDCClientList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCClientSpec) DeepCopyInto(out *OIDCClientSpec) {
	*out = *in
	if in.AllowedRedirectURIs != nil {
		in, out := &in.AllowedRedirectURIs, &out.AllowedRedirectURIs
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowedGrantTypes != nil {
		in, out := &in.AllowedGrantTypes, &out.AllowedGrantTypes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowedScopes != nil {
		in, out := &in.AllowedScopes, &out.AllowedScopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCClientSpec.
func (in *OIDCClientSpec) DeepCopy() *OIDCClientSpec {
	if in == nil {
		return nil
	}
	out := new(OIDCClientSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *OIDCClientStatus) DeepCopyInto(out *OIDCClientStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new OIDCClientStatus.
func (in *OIDCClientStatus) DeepCopy() *OIDCClientStatus {
	if in == nil {
		return nil
	}
	out := new(OIDCClientStatus)
	in.DeepCopyInto(out)
	return out
}
