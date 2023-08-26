//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Backend) DeepCopyInto(out *Backend) {
	*out = *in
	if in.MaxRequests != nil {
		in, out := &in.MaxRequests, &out.MaxRequests
		*out = new(uint32)
		**out = **in
	}
	if in.Interceptors != nil {
		in, out := &in.Interceptors, &out.Interceptors
		*out = make([]Interceptor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceRef)
		**out = **in
	}
	if in.Localities != nil {
		in, out := &in.Localities, &out.Localities
		*out = make([]Locality, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Backend.
func (in *Backend) DeepCopy() *Backend {
	if in == nil {
		return nil
	}
	out := new(Backend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaultAbort) DeepCopyInto(out *FaultAbort) {
	*out = *in
	if in.Status != nil {
		in, out := &in.Status, &out.Status
		*out = new(uint32)
		**out = **in
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = new(MetadataFault)
		**out = **in
	}
	if in.Percentage != nil {
		in, out := &in.Percentage, &out.Percentage
		*out = new(Fraction)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaultAbort.
func (in *FaultAbort) DeepCopy() *FaultAbort {
	if in == nil {
		return nil
	}
	out := new(FaultAbort)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaultDelay) DeepCopyInto(out *FaultDelay) {
	*out = *in
	if in.Fixed != nil {
		in, out := &in.Fixed, &out.Fixed
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = new(MetadataFault)
		**out = **in
	}
	if in.Percentage != nil {
		in, out := &in.Percentage, &out.Percentage
		*out = new(Fraction)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaultDelay.
func (in *FaultDelay) DeepCopy() *FaultDelay {
	if in == nil {
		return nil
	}
	out := new(FaultDelay)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FaultInterceptor) DeepCopyInto(out *FaultInterceptor) {
	*out = *in
	if in.Delay != nil {
		in, out := &in.Delay, &out.Delay
		*out = new(FaultDelay)
		(*in).DeepCopyInto(*out)
	}
	if in.Abort != nil {
		in, out := &in.Abort, &out.Abort
		*out = new(FaultAbort)
		(*in).DeepCopyInto(*out)
	}
	if in.MaxActiveFaults != nil {
		in, out := &in.MaxActiveFaults, &out.MaxActiveFaults
		*out = new(uint32)
		**out = **in
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make([]HeaderMatcher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FaultInterceptor.
func (in *FaultInterceptor) DeepCopy() *FaultInterceptor {
	if in == nil {
		return nil
	}
	out := new(FaultInterceptor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Fraction) DeepCopyInto(out *Fraction) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Fraction.
func (in *Fraction) DeepCopy() *Fraction {
	if in == nil {
		return nil
	}
	out := new(Fraction)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GRPCListener) DeepCopyInto(out *GRPCListener) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GRPCListener.
func (in *GRPCListener) DeepCopy() *GRPCListener {
	if in == nil {
		return nil
	}
	out := new(GRPCListener)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GRPCListener) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GRPCListenerList) DeepCopyInto(out *GRPCListenerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]GRPCListener, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GRPCListenerList.
func (in *GRPCListenerList) DeepCopy() *GRPCListenerList {
	if in == nil {
		return nil
	}
	out := new(GRPCListenerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *GRPCListenerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GRPCListenerSpec) DeepCopyInto(out *GRPCListenerSpec) {
	*out = *in
	if in.MaxStreamDuration != nil {
		in, out := &in.MaxStreamDuration, &out.MaxStreamDuration
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Interceptors != nil {
		in, out := &in.Interceptors, &out.Interceptors
		*out = make([]Interceptor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Routes != nil {
		in, out := &in.Routes, &out.Routes
		*out = make([]Route, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GRPCListenerSpec.
func (in *GRPCListenerSpec) DeepCopy() *GRPCListenerSpec {
	if in == nil {
		return nil
	}
	out := new(GRPCListenerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HeaderMatcher) DeepCopyInto(out *HeaderMatcher) {
	*out = *in
	if in.Exact != nil {
		in, out := &in.Exact, &out.Exact
		*out = new(string)
		**out = **in
	}
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = new(RegexMatcher)
		**out = **in
	}
	if in.Range != nil {
		in, out := &in.Range, &out.Range
		*out = new(RangeMatcher)
		**out = **in
	}
	if in.Present != nil {
		in, out := &in.Present, &out.Present
		*out = new(bool)
		**out = **in
	}
	if in.Prefix != nil {
		in, out := &in.Prefix, &out.Prefix
		*out = new(string)
		**out = **in
	}
	if in.Suffix != nil {
		in, out := &in.Suffix, &out.Suffix
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HeaderMatcher.
func (in *HeaderMatcher) DeepCopy() *HeaderMatcher {
	if in == nil {
		return nil
	}
	out := new(HeaderMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Interceptor) DeepCopyInto(out *Interceptor) {
	*out = *in
	if in.Fault != nil {
		in, out := &in.Fault, &out.Fault
		*out = new(FaultInterceptor)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Interceptor.
func (in *Interceptor) DeepCopy() *Interceptor {
	if in == nil {
		return nil
	}
	out := new(Interceptor)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Locality) DeepCopyInto(out *Locality) {
	*out = *in
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceRef)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Locality.
func (in *Locality) DeepCopy() *Locality {
	if in == nil {
		return nil
	}
	out := new(Locality)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetadataFault) DeepCopyInto(out *MetadataFault) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetadataFault.
func (in *MetadataFault) DeepCopy() *MetadataFault {
	if in == nil {
		return nil
	}
	out := new(MetadataFault)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MetadataMatcher) DeepCopyInto(out *MetadataMatcher) {
	*out = *in
	if in.Exact != nil {
		in, out := &in.Exact, &out.Exact
		*out = new(string)
		**out = **in
	}
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = new(RegexMatcher)
		**out = **in
	}
	if in.Range != nil {
		in, out := &in.Range, &out.Range
		*out = new(RangeMatcher)
		**out = **in
	}
	if in.Present != nil {
		in, out := &in.Present, &out.Present
		*out = new(bool)
		**out = **in
	}
	if in.Prefix != nil {
		in, out := &in.Prefix, &out.Prefix
		*out = new(string)
		**out = **in
	}
	if in.Suffix != nil {
		in, out := &in.Suffix, &out.Suffix
		*out = new(string)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MetadataMatcher.
func (in *MetadataMatcher) DeepCopy() *MetadataMatcher {
	if in == nil {
		return nil
	}
	out := new(MetadataMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MethodMatcher) DeepCopyInto(out *MethodMatcher) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MethodMatcher.
func (in *MethodMatcher) DeepCopy() *MethodMatcher {
	if in == nil {
		return nil
	}
	out := new(MethodMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PathMatcher) DeepCopyInto(out *PathMatcher) {
	*out = *in
	if in.Regex != nil {
		in, out := &in.Regex, &out.Regex
		*out = new(RegexMatcher)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PathMatcher.
func (in *PathMatcher) DeepCopy() *PathMatcher {
	if in == nil {
		return nil
	}
	out := new(PathMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortRef) DeepCopyInto(out *PortRef) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortRef.
func (in *PortRef) DeepCopy() *PortRef {
	if in == nil {
		return nil
	}
	out := new(PortRef)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RangeMatcher) DeepCopyInto(out *RangeMatcher) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RangeMatcher.
func (in *RangeMatcher) DeepCopy() *RangeMatcher {
	if in == nil {
		return nil
	}
	out := new(RangeMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RegexMatcher) DeepCopyInto(out *RegexMatcher) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RegexMatcher.
func (in *RegexMatcher) DeepCopy() *RegexMatcher {
	if in == nil {
		return nil
	}
	out := new(RegexMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Route) DeepCopyInto(out *Route) {
	*out = *in
	if in.Matcher != nil {
		in, out := &in.Matcher, &out.Matcher
		*out = new(RouteMatcher)
		(*in).DeepCopyInto(*out)
	}
	if in.Interceptors != nil {
		in, out := &in.Interceptors, &out.Interceptors
		*out = make([]Interceptor, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.MaxStreamDuration != nil {
		in, out := &in.MaxStreamDuration, &out.MaxStreamDuration
		*out = new(v1.Duration)
		**out = **in
	}
	if in.GrpcTimeoutHeaderMax != nil {
		in, out := &in.GrpcTimeoutHeaderMax, &out.GrpcTimeoutHeaderMax
		*out = new(v1.Duration)
		**out = **in
	}
	if in.Backends != nil {
		in, out := &in.Backends, &out.Backends
		*out = make([]Backend, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Route.
func (in *Route) DeepCopy() *Route {
	if in == nil {
		return nil
	}
	out := new(Route)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RouteMatcher) DeepCopyInto(out *RouteMatcher) {
	*out = *in
	if in.Method != nil {
		in, out := &in.Method, &out.Method
		*out = new(MethodMatcher)
		**out = **in
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(ServiceMatcher)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make([]MetadataMatcher, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Fraction != nil {
		in, out := &in.Fraction, &out.Fraction
		*out = new(Fraction)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RouteMatcher.
func (in *RouteMatcher) DeepCopy() *RouteMatcher {
	if in == nil {
		return nil
	}
	out := new(RouteMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceMatcher) DeepCopyInto(out *ServiceMatcher) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceMatcher.
func (in *ServiceMatcher) DeepCopy() *ServiceMatcher {
	if in == nil {
		return nil
	}
	out := new(ServiceMatcher)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceRef) DeepCopyInto(out *ServiceRef) {
	*out = *in
	out.Port = in.Port
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceRef.
func (in *ServiceRef) DeepCopy() *ServiceRef {
	if in == nil {
		return nil
	}
	out := new(ServiceRef)
	in.DeepCopyInto(out)
	return out
}
