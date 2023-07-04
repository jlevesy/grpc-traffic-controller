package v1alpha1

type RegexMatcher struct {
	// Regexp to evaluate the path against.
	Regex string `json:"regex,omitempty"`
	// The regexp engine to use.
	// +kubebuilder:validation:Enum:=re2
	// +kubebuilder:default:=re2
	Engine string `json:"engine,omitempty"`
}

type RangeMatcher struct {
	// Start of the range (inclusive)
	Start int64 `json:"start,omitempty"`
	// End of the range (exclusive)
	End int64 `json:"end,omitempty"`
}

// PathMatcher indicates a match based on the path of a gRPC call.
type PathMatcher struct {
	// Path Must match the prefix of the request.
	// +optional
	// +kubebuilder:default:=/
	Prefix string `json:"prefix,omitempty"`
	// Path Must match exactly.
	// +optional
	Path string `json:"path,omitempty"`
	// Path Must Match a Regex.
	// +optional
	Regex *RegexMatcher `json:"regex,omitempty"`
}

// HeaderMatcher indicates a match based on an http header.
type HeaderMatcher struct {
	// Name of the header to match.
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

type Fraction struct {
	// Numerator of the fraction
	Numerator uint32 `json:"numerator,omitempty"`
	// Denominator of the fration.
	// +kubebuilder:validation:Enum:=hundred;ten_thousand;million
	// +kubebuilder:default:=hundred
	Denominator string `json:"denominator,omitempty"`
}
