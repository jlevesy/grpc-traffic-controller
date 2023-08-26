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

// RetryBackoffApplyConfiguration represents an declarative configuration of the RetryBackoff type for use
// with apply.
type RetryBackoffApplyConfiguration struct {
	BaseInterval *v1.Duration `json:"baseInterval,omitempty"`
	MaxInterval  *v1.Duration `json:"maxInterval,omitempty"`
}

// RetryBackoffApplyConfiguration constructs an declarative configuration of the RetryBackoff type for use with
// apply.
func RetryBackoff() *RetryBackoffApplyConfiguration {
	return &RetryBackoffApplyConfiguration{}
}

// WithBaseInterval sets the BaseInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the BaseInterval field is set to the value of the last call.
func (b *RetryBackoffApplyConfiguration) WithBaseInterval(value v1.Duration) *RetryBackoffApplyConfiguration {
	b.BaseInterval = &value
	return b
}

// WithMaxInterval sets the MaxInterval field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the MaxInterval field is set to the value of the last call.
func (b *RetryBackoffApplyConfiguration) WithMaxInterval(value v1.Duration) *RetryBackoffApplyConfiguration {
	b.MaxInterval = &value
	return b
}
