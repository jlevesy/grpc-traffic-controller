package kxds

import (
	"errors"

	faultv31 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	faultv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	kxdsv1alpha1 "github.com/jlevesy/grpc-traffic-controller/api/kxds/v1alpha1"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func makeFilters(filters []kxdsv1alpha1.Filter) ([]*hcm.HttpFilter, error) {
	routerFilter := &hcm.HttpFilter{
		Name: wellknown.Router,
		ConfigType: &hcm.HttpFilter_TypedConfig{
			TypedConfig: mustAny(&router.Router{}),
		},
	}

	if len(filters) == 0 {
		return []*hcm.HttpFilter{
			routerFilter,
		}, nil
	}

	hcmFilters := make([]*hcm.HttpFilter, len(filters)+1)

	for i, filterSpec := range filters {
		var err error

		hcmFilters[i], err = makeFilter(filterSpec)
		if err != nil {
			return nil, err
		}
	}

	// Always set the router last.
	hcmFilters[len(filters)] = routerFilter

	return hcmFilters, nil
}

func makeFilter(filter kxdsv1alpha1.Filter) (*hcm.HttpFilter, error) {
	switch {
	case filter.Fault != nil:
		faultFilter, err := makeFaultFilter(filter.Fault)
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

func makeFaultFilter(f *kxdsv1alpha1.FaultFilter) (*faultv3.HTTPFault, error) {
	var ff faultv3.HTTPFault

	if f.Delay != nil {
		ff.Delay = &faultv31.FaultDelay{}

		switch {
		case f.Delay.Fixed != nil:
			ff.Delay.FaultDelaySecifier = &faultv31.FaultDelay_FixedDelay{
				FixedDelay: durationpb.New(f.Delay.Fixed.Duration),
			}
		case f.Delay.Header != nil:
			ff.Delay.FaultDelaySecifier = &faultv31.FaultDelay_HeaderDelay_{}
		default:
			return nil, errors.New("malformed delay fault filter")
		}

		if f.Delay.Percentage != nil {
			var err error

			ff.Delay.Percentage, err = makeFractionalPercent(f.Delay.Percentage)

			if err != nil {
				return nil, err
			}
		}
	}

	if f.Abort != nil {
		ff.Abort = &faultv3.FaultAbort{}

		switch {
		case f.Abort.HTTPStatus != nil:
			ff.Abort.ErrorType = &faultv3.FaultAbort_HttpStatus{
				HttpStatus: *f.Abort.HTTPStatus,
			}
		case f.Abort.GRPCStatus != nil:
			ff.Abort.ErrorType = &faultv3.FaultAbort_GrpcStatus{
				GrpcStatus: *f.Abort.GRPCStatus,
			}
		case f.Abort.Header != nil:
			ff.Abort.ErrorType = &faultv3.FaultAbort_HeaderAbort_{}
		default:
			return nil, errors.New("malformed abort fault filter")
		}

		if f.Abort.Percentage != nil {
			var err error

			ff.Abort.Percentage, err = makeFractionalPercent(f.Abort.Percentage)

			if err != nil {
				return nil, err
			}
		}
	}

	if f.MaxActiveFaults != nil {
		ff.MaxActiveFaults = wrapperspb.UInt32(*f.MaxActiveFaults)
	}

	return &ff, nil
}
