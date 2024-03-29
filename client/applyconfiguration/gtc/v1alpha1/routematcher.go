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

// RouteMatcherApplyConfiguration represents an declarative configuration of the RouteMatcher type for use
// with apply.
type RouteMatcherApplyConfiguration struct {
	Method    *MethodMatcherApplyConfiguration    `json:"method,omitempty"`
	Service   *ServiceMatcherApplyConfiguration   `json:"service,omitempty"`
	Namespace *string                             `json:"namespace,omitempty"`
	Metadata  []MetadataMatcherApplyConfiguration `json:"metadata,omitempty"`
	Fraction  *FractionApplyConfiguration         `json:"fraction,omitempty"`
}

// RouteMatcherApplyConfiguration constructs an declarative configuration of the RouteMatcher type for use with
// apply.
func RouteMatcher() *RouteMatcherApplyConfiguration {
	return &RouteMatcherApplyConfiguration{}
}

// WithMethod sets the Method field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Method field is set to the value of the last call.
func (b *RouteMatcherApplyConfiguration) WithMethod(value *MethodMatcherApplyConfiguration) *RouteMatcherApplyConfiguration {
	b.Method = value
	return b
}

// WithService sets the Service field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Service field is set to the value of the last call.
func (b *RouteMatcherApplyConfiguration) WithService(value *ServiceMatcherApplyConfiguration) *RouteMatcherApplyConfiguration {
	b.Service = value
	return b
}

// WithNamespace sets the Namespace field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Namespace field is set to the value of the last call.
func (b *RouteMatcherApplyConfiguration) WithNamespace(value string) *RouteMatcherApplyConfiguration {
	b.Namespace = &value
	return b
}

// WithMetadata adds the given value to the Metadata field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Metadata field.
func (b *RouteMatcherApplyConfiguration) WithMetadata(values ...*MetadataMatcherApplyConfiguration) *RouteMatcherApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithMetadata")
		}
		b.Metadata = append(b.Metadata, *values[i])
	}
	return b
}

// WithFraction sets the Fraction field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Fraction field is set to the value of the last call.
func (b *RouteMatcherApplyConfiguration) WithFraction(value *FractionApplyConfiguration) *RouteMatcherApplyConfiguration {
	b.Fraction = value
	return b
}
