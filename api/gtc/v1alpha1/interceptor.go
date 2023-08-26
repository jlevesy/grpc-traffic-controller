package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type Interceptor struct {
	// Fault Interceptor configuration.
	// +optional
	Fault *FaultInterceptor `json:"fault,omitempty"`
}

type FaultInterceptor struct {
	// Inject a delay.
	Delay *FaultDelay `json:"delay,omitempty"`
	// Abort the call.
	Abort *FaultAbort `json:"abort,omitempty"`
	// The maximum number of faults that can be active at a single time.
	MaxActiveFaults *uint32 `json:"maxActiveFaults,omitempty"`
	// Specifies a set of headers that the filter should match on.
	Headers []HeaderMatcher `json:"headers,omitempty"`
}

type FaultAbort struct {
	// Returns the gRPC status code.
	Code *uint32 `json:"code,omitempty"`
	// Metadata adds a fault controlled by an call metadata.
	Metadata *MetadataFault `json:"metadata,omitempty"`
	// Percentage controls how much this fault will occur.
	Percentage *Fraction `json:"percentage,omitempty"`
}

type MetadataFault struct{}

type FaultDelay struct {
	// FixedDelay adds a fixed delay before a call.
	Fixed *metav1.Duration `json:"fixed,omitempty"`
	// Metadata adds a fault controlled by an call metadata.
	Metadata *MetadataFault `json:"metadata,omitempty"`
	// Percentage controls how much this fault will occur.
	Percentage *Fraction `json:"percentage,omitempty"`
}
