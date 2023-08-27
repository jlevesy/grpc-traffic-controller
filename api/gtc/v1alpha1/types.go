/*
Copyright 2022.

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

package v1alpha1

import (
	"path"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GRPCListener is the Schema for the services API
// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type GRPCListener struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GRPCListenerSpec `json:"spec,omitempty"`
}

// GRPCListenerSpec defines the desired state of Service
type GRPCListenerSpec struct {
	// MaxStreamDuration is the total duration to keep alive an HTTP request/response stream.
	// If the time limit is reached the stream will be reset independent of any other timeouts.
	// If not specified, this value is not set.
	// +optional
	MaxStreamDuration *metav1.Duration `json:"maxStreamDuration,omitempty"`
	// Interceptors represent the list of interceptors applied globally in this listener.
	// +optional
	Interceptors []Interceptor `json:"interceptors,omitempty"`
	// Retry indicates a retry policy to be applied on every route of this listener.
	// +optional
	Retry *RetryPolicy `json:"retry,omitempty"`
	// Routes lists all the routes defined for an GRPCListener.
	Routes []Route `json:"routes,omitempty"`
}

// Route allows to match an outoing request to a specific cluster, it allows to do HTTP level manipulation on the outgoing requests as well as matching.
type Route struct {
	// Matcher define a way of matching a specific route.
	Matcher *RouteMatcher `json:"matcher,omitempty"`

	// Interceptors are a list of interceptor overrides to apply to this route.
	// Note that the interceptors defined here must me also defined at the listener level.
	Interceptors []Interceptor `json:"interceptors,omitempty"`

	// HashPolicy are a list of heuristics to apply to obtain a hash for a given request.
	// Multiple policies result are combined.
	HashPolicy []HashPolicy `json:"hashPolicy,omitempty"`

	// Only handle a fraction of matching requests.
	// RuntimeFraction *Fraction `json:"fraction,omitempty"`
	// Specifies the maximum duration allowed for streams on the route.
	MaxStreamDuration *metav1.Duration `json:"maxStreamDuration,omitempty"`

	// Specifies the maximum duration allowed for streams on the route.
	// If present, and the request contains a `grpc-timeout header
	// <https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md>`_, use that value as the
	// *max_stream_duration*, but limit the applied timeout to the maximum value specified here.
	// If set to 0, the `grpc-timeout` header is used without modification.
	GrpcTimeoutHeaderMax *metav1.Duration `json:"grpcTimeoutHeaderMax,omitempty"`

	// Retry indicates a retry policy to be applied for this route.
	// +optional
	Retry *RetryPolicy `json:"retry,omitempty"`

	// Backends is the list of all backends serving that route.
	Backends []Backend `json:"backends,omitempty"`
}

type RouteMatcher struct {
	// Method allows to match a specific method of a grpc service.
	Method *MethodMatcher `json:"method,omitempty"`

	// Service allows to match a specific service.
	Service *ServiceMatcher `json:"service,omitempty"`

	// Namespace allows to match a specific namespace.
	Namespace *string `json:"namespace,omitempty"`

	// Metadata allows to match on a specific set of call metadata.
	Metadata []MetadataMatcher `json:"metadata,omitempty"`

	// Fraction allows to match a certain percentage of calls.
	Fraction *Fraction `json:"fraction,omitempty"`
}

type MetadataMatcher struct {
	// Name of the metadata to match.
	Name string `json:"name,omitempty"`
	// Match the exact value of a header.
	Exact *string `json:"exact,omitempty"`
	// Match a regex. Must match the whole value.
	Regex *RegexMatcher `json:"regex,omitempty"`
	// Header Value must match a range.
	Range *RangeMatcher `json:"range,omitempty"`
	// Header must be present.
	Present *bool `json:"present,omitempty"`
	// Header value must have a prefix.
	Prefix *string `json:"prefix,omitempty"`
	// Header value must have a suffix.
	Suffix *string `json:"suffix,omitempty"`
	// Invert that header match.
	Invert bool `json:"invert,omitempty"`
}

type MethodMatcher struct {
	Namespace string `json:"namespace,omitempty"`
	Service   string `json:"service,omitempty"`
	Method    string `json:"method,omitempty"`
}

func (mm *MethodMatcher) Path() string {
	return "/" + path.Join(
		mm.Namespace+"."+mm.Service,
		mm.Method,
	)
}

type ServiceMatcher struct {
	Namespace string `json:"namespace,omitempty"`
	Service   string `json:"service,omitempty"`
}

func (sm *ServiceMatcher) Prefix() string {
	return "/" + sm.Namespace + "." + sm.Service
}

// Backend is a group of backend servers serving the same services.
type Backend struct {
	// Weight is the weight of this cluster.
	// +optional
	// +kubebuilder:default:=1
	Weight uint32 `json:"weight,omitempty"`
	// MaxRequests qualifies the maximum number of parallel requests allowd to the upstream cluster.
	MaxRequests *uint32 `json:"maxRequests,omitempty"`

	// Weight is the weight of this cluster.
	// +optional
	// +kubebuilder:validation:Enum:=round_robin;roundRobin;ring_hash;ringHash
	// +kubebuilder:default:=round_robin
	LBPolicy string `json:"lbPolicy,omitempty"`

	//RingHashConfig is an optional configuration for the ring_hash lb policy
	// +optional
	RingHashConfig *RingHashConfig `json:"ringHashConfig"`

	// Interceptors are a list of interceptor overrides to apply to this backend.
	// Note that the interceptors defined here must me also defined at the listener level.
	Interceptors []Interceptor `json:"interceptors,omitempty"`

	// Service is a reference to a k8s service.
	// +optional
	Service *ServiceRef `json:"service,omitempty"`
	// Localities is a list of prioritized and weighted localities for a backend.
	// +optional
	Localities []Locality `json:"localities,omitempty"`
}

type RingHashConfig struct {
	// Minimum hash ring size. The larger the ring is (that is, the more hashes there are for each
	// provided host) the better the request distribution will reflect the desired weights.
	// +kubebuilder:default:=1024
	MinRingSize uint64 `json:"minRingSize,omitempty"`
	// Maximum hash ring size. Defaults to 8M entries, and limited to 8M entries, but can be lowered
	// to further constrain resource use.
	// +kubebuilder:default:=838860
	MaxRingSize uint64 `json:"maxRingSize,omitempty"`
}

// Locality is a weighted and prioritized locality for a backend.
type Locality struct {
	// Weight of the locality, defaults to one.
	// +optional
	// +kubebuilder:default:=1
	Weight uint32 `json:"weight,omitempty"`
	// Priority of the locality, if defined, all entries must unique for a given priority and priority should be defined without any gap.
	// +optional
	Priority uint32 `json:"priority,omitempty"`
	// Service is a reference to a kubernetes service.
	// +optional
	Service *ServiceRef `json:"service,omitempty"`
}

// PortRef represents a reference to a port. This could be done either by number or by name.
// +kubebuilder:validation:MaxProperties:=1
type PortRef struct {
	// +optional
	Number int32 `json:"number,omitempty"`
	// +optional
	Name string `json:"name,omitempty"`
}

// ServiceRef is a reference to kubernetes service.
type ServiceRef struct {
	Name string `json:"name,omitempty"`
	// +optional
	Namespace string  `json:"namespace,omitempty"`
	Port      PortRef `json:"port,omitempty"`
}

// RetryPolicy indicates a retry policy.
type RetryPolicy struct {
	// Specifies the conditions under which retry takes place.
	RetryOn []string `json:"retryOn,omitempty"`
	// Specifies the allowed number of retries. This parameter is optional and defaults to 1.
	// +optional
	// +kubebuilder:default:=1
	NumRetries *uint32 `json:"numRetries,omitempty"`
	//  Specifies parameters that control exponential retry back off. This parameter is optional, in which case the default base interval is 25 milliseconds
	// +optional
	Backoff *RetryBackoff `json:"backoff,omitempty"`
}

// RetryBackoff indicates a backoff retry policy.
type RetryBackoff struct {
	// Specifies the base interval between retries.
	// This parameter is required and must be greater than zero. Values less than 1 ms are rounded up to 1 ms.
	BaseInterval metav1.Duration `json:"baseInterval,omitempty"`
	// Specifies the maximum interval between retries. This parameter is optional, but must be greater than or equal to the base_interval if set. The default is 10 times the base_interval
	// +optional
	MaxInterval *metav1.Duration `json:"maxInterval,omitempty"`
}

// HashPolicy indicates a way of obtaining a hash from a request.
// It could be either using a medatada name
// be based on the channel ID of the request.
type HashPolicy struct {
	// Metadata indicates rpc metadata call value to obtain a hash.
	Metadata string `json:"metadata,omitempty"`
	// Channel indicates to use the chanel_id to obtain a hash.
	Channel *bool `json:"channel,omitempty"`
	// Terminal tells to stop the hashing process if this policy is successful.
	Terminal bool `json:"terminal,omitempty"`
}

// GRPCListenerList contains a list of GRPCListener
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GRPCListenerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GRPCListener `json:"items"`
}
