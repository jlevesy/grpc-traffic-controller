package gtc

import (
	"errors"

	faultv31 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	faultv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	gtcv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/gtc/v1alpha1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func makeFilters(interceptors []gtcv1alpha1.Interceptor) ([]*hcm.HttpFilter, error) {
	routerFilter := &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: mustAny(&router.Router{}),
		},
	}

	if len(interceptors) == 0 {
		return []*hcm.HttpFilter{
			routerFilter,
		}, nil
	}

	hcmFilters := make([]*hcm.HttpFilter, len(interceptors)+1)

	for i, interceptorSpec := range interceptors {
		var err error

		hcmFilters[i], err = makeFilter(interceptorSpec)
		if err != nil {
			return nil, err
		}
	}

	// Always set the router last.
	hcmFilters[len(interceptors)] = routerFilter

	return hcmFilters, nil
}

func makeFilter(interceptor gtcv1alpha1.Interceptor) (*hcm.HttpFilter, error) {
	switch {
	case interceptor.Fault != nil:
		faultFilter, err := makeFaultFilter(interceptor.Fault)
		if err != nil {
			return nil, err
		}

		return &hcm.HttpFilter{
			Name: wellknown.Fault,
			ConfigType: &hcm.HttpFilter_TypedConfig{
				TypedConfig: mustAny(faultFilter),
			},
		}, nil
	default:
		return nil, errors.New("malformed filter")
	}
}

func makeFaultFilter(fault *gtcv1alpha1.FaultInterceptor) (*faultv3.HTTPFault, error) {
	var ff faultv3.HTTPFault

	if fault.Delay != nil {
		ff.Delay = &faultv31.FaultDelay{}

		switch {
		case fault.Delay.Fixed != nil:
			ff.Delay.FaultDelaySecifier = &faultv31.FaultDelay_FixedDelay{
				FixedDelay: durationpb.New(fault.Delay.Fixed.Duration),
			}
		case fault.Delay.Metadata != nil:
			ff.Delay.FaultDelaySecifier = &faultv31.FaultDelay_HeaderDelay_{}
		default:
			return nil, errors.New("malformed delay fault filter")
		}

		if fault.Delay.Percentage != nil {
			var err error

			ff.Delay.Percentage, err = makeFractionalPercent(fault.Delay.Percentage)

			if err != nil {
				return nil, err
			}
		}
	}

	if fault.Abort != nil {
		ff.Abort = &faultv3.FaultAbort{}

		switch {
		case fault.Abort.Status != nil:
			ff.Abort.ErrorType = &faultv3.FaultAbort_GrpcStatus{
				GrpcStatus: *fault.Abort.Status,
			}
		case fault.Abort.Metadata != nil:
			ff.Abort.ErrorType = &faultv3.FaultAbort_HeaderAbort_{}
		default:
			return nil, errors.New("malformed abort fault filter")
		}

		if fault.Abort.Percentage != nil {
			var err error

			ff.Abort.Percentage, err = makeFractionalPercent(fault.Abort.Percentage)

			if err != nil {
				return nil, err
			}
		}
	}

	if fault.MaxActiveFaults != nil {
		ff.MaxActiveFaults = wrapperspb.UInt32(*fault.MaxActiveFaults)
	}

	return &ff, nil
}
