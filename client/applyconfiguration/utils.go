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

package applyconfiguration

import (
	v1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/client/applyconfiguration/gtc/v1alpha1"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
)

// ForKind returns an apply configuration type for the given GroupVersionKind, or nil if no
// apply configuration type exists for the given GroupVersionKind.
func ForKind(kind schema.GroupVersionKind) interface{} {
	switch kind {
	// Group=api.gtc.dev, Version=v1alpha1
	case v1alpha1.SchemeGroupVersion.WithKind("Cluster"):
		return &gtcv1alpha1.ClusterApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("ClusterRef"):
		return &gtcv1alpha1.ClusterRefApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("DefaultCluster"):
		return &gtcv1alpha1.DefaultClusterApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("FaultAbort"):
		return &gtcv1alpha1.FaultAbortApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("FaultDelay"):
		return &gtcv1alpha1.FaultDelayApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("FaultFilter"):
		return &gtcv1alpha1.FaultFilterApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Filter"):
		return &gtcv1alpha1.FilterApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Fraction"):
		return &gtcv1alpha1.FractionApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("GRPCListener"):
		return &gtcv1alpha1.GRPCListenerApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("GRPCListenerSpec"):
		return &gtcv1alpha1.GRPCListenerSpecApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("HeaderMatcher"):
		return &gtcv1alpha1.HeaderMatcherApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Locality"):
		return &gtcv1alpha1.LocalityApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("PathMatcher"):
		return &gtcv1alpha1.PathMatcherApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("PortRef"):
		return &gtcv1alpha1.PortRefApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("RangeMatcher"):
		return &gtcv1alpha1.RangeMatcherApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("RegexMatcher"):
		return &gtcv1alpha1.RegexMatcherApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("Route"):
		return &gtcv1alpha1.RouteApplyConfiguration{}
	case v1alpha1.SchemeGroupVersion.WithKind("ServiceRef"):
		return &gtcv1alpha1.ServiceRefApplyConfiguration{}

	}
	return nil
}
