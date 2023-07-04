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

// HeaderMatcherApplyConfiguration represents an declarative configuration of the HeaderMatcher type for use
// with apply.
type HeaderMatcherApplyConfiguration struct {
	Name    *string                         `json:"name,omitempty"`
	Exact   *string                         `json:"exact,omitempty"`
	Regex   *RegexMatcherApplyConfiguration `json:"regex,omitempty"`
	Range   *RangeMatcherApplyConfiguration `json:"range,omitempty"`
	Present *bool                           `json:"present,omitempty"`
	Prefix  *string                         `json:"prefix,omitempty"`
	Suffix  *string                         `json:"suffix,omitempty"`
	Invert  *bool                           `json:"invert,omitempty"`
}

// HeaderMatcherApplyConfiguration constructs an declarative configuration of the HeaderMatcher type for use with
// apply.
func HeaderMatcher() *HeaderMatcherApplyConfiguration {
	return &HeaderMatcherApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithName(value string) *HeaderMatcherApplyConfiguration {
	b.Name = &value
	return b
}

// WithExact sets the Exact field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Exact field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithExact(value string) *HeaderMatcherApplyConfiguration {
	b.Exact = &value
	return b
}

// WithRegex sets the Regex field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Regex field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithRegex(value *RegexMatcherApplyConfiguration) *HeaderMatcherApplyConfiguration {
	b.Regex = value
	return b
}

// WithRange sets the Range field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Range field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithRange(value *RangeMatcherApplyConfiguration) *HeaderMatcherApplyConfiguration {
	b.Range = value
	return b
}

// WithPresent sets the Present field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Present field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithPresent(value bool) *HeaderMatcherApplyConfiguration {
	b.Present = &value
	return b
}

// WithPrefix sets the Prefix field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Prefix field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithPrefix(value string) *HeaderMatcherApplyConfiguration {
	b.Prefix = &value
	return b
}

// WithSuffix sets the Suffix field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Suffix field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithSuffix(value string) *HeaderMatcherApplyConfiguration {
	b.Suffix = &value
	return b
}

// WithInvert sets the Invert field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Invert field is set to the value of the last call.
func (b *HeaderMatcherApplyConfiguration) WithInvert(value bool) *HeaderMatcherApplyConfiguration {
	b.Invert = &value
	return b
}
