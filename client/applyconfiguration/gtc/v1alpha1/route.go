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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RouteApplyConfiguration represents an declarative configuration of the Route type for use
// with apply.
type RouteApplyConfiguration struct {
	Path                 *PathMatcherApplyConfiguration    `json:"path,omitempty"`
	Headers              []HeaderMatcherApplyConfiguration `json:"headers,omitempty"`
	CaseSensitive        *bool                             `json:"caseSensitive,omitempty"`
	RuntimeFraction      *FractionApplyConfiguration       `json:"fraction,omitempty"`
	MaxStreamDuration    *v1.Duration                      `json:"maxStreamDuration,omitempty"`
	GrpcTimeoutHeaderMax *v1.Duration                      `json:"grpcTimeoutHeaderMax,omitempty"`
	Clusters             []ClusterRefApplyConfiguration    `json:"clusters,omitempty"`
}

// RouteApplyConfiguration constructs an declarative configuration of the Route type for use with
// apply.
func Route() *RouteApplyConfiguration {
	return &RouteApplyConfiguration{}
}

// WithPath sets the Path field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Path field is set to the value of the last call.
func (b *RouteApplyConfiguration) WithPath(value *PathMatcherApplyConfiguration) *RouteApplyConfiguration {
	b.Path = value
	return b
}

// WithHeaders adds the given value to the Headers field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Headers field.
func (b *RouteApplyConfiguration) WithHeaders(values ...*HeaderMatcherApplyConfiguration) *RouteApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithHeaders")
		}
		b.Headers = append(b.Headers, *values[i])
	}
	return b
}

// WithCaseSensitive sets the CaseSensitive field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CaseSensitive field is set to the value of the last call.
func (b *RouteApplyConfiguration) WithCaseSensitive(value bool) *RouteApplyConfiguration {
	b.CaseSensitive = &value
	return b
}

// WithRuntimeFraction sets the RuntimeFraction field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RuntimeFraction field is set to the value of the last call.
func (b *RouteApplyConfiguration) WithRuntimeFraction(value *FractionApplyConfiguration) *RouteApplyConfiguration {
	b.RuntimeFraction = value
	return b
}

// WithMaxStreamDuration sets the MaxStreamDuration field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MaxStreamDuration field is set to the value of the last call.
func (b *RouteApplyConfiguration) WithMaxStreamDuration(value v1.Duration) *RouteApplyConfiguration {
	b.MaxStreamDuration = &value
	return b
}

// WithGrpcTimeoutHeaderMax sets the GrpcTimeoutHeaderMax field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GrpcTimeoutHeaderMax field is set to the value of the last call.
func (b *RouteApplyConfiguration) WithGrpcTimeoutHeaderMax(value v1.Duration) *RouteApplyConfiguration {
	b.GrpcTimeoutHeaderMax = &value
	return b
}

// WithClusters adds the given value to the Clusters field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Clusters field.
func (b *RouteApplyConfiguration) WithClusters(values ...*ClusterRefApplyConfiguration) *RouteApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithClusters")
		}
		b.Clusters = append(b.Clusters, *values[i])
	}
	return b
}
