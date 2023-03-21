package kxds

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"

	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	anyv1 "github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kxdsv1alpha1 "github.com/jlevesy/kxds/api/kxds/v1alpha1"
)

func makeDuration(duration *kmetav1.Duration) *durationpb.Duration {
	if duration == nil {
		return nil
	}

	return durationpb.New(duration.Duration)
}

func mustAny(msg protoreflect.ProtoMessage) *anypb.Any {
	p, err := anypb.New(msg)
	if err != nil {
		panic(err)
	}

	return p
}

func makeFractionalPercent(p *kxdsv1alpha1.Fraction) (*typev3.FractionalPercent, error) {
	denominator, ok := typev3.FractionalPercent_DenominatorType_value[strings.ToUpper(p.Denominator)]
	if !ok {
		return nil, fmt.Errorf(
			"unsupported denominator %q for runtime fraction",
			p.Denominator,
		)
	}

	return &typev3.FractionalPercent{
		Numerator:   p.Numerator,
		Denominator: typev3.FractionalPercent_DenominatorType(denominator),
	}, nil
}

func computeVersionInfo(vs []string) (string, error) {
	hasher := sha256.New()

	for _, v := range vs {
		_, err := hasher.Write([]byte(v))
		if err != nil {
			return "", err
		}
	}

	return base64.StdEncoding.EncodeToString(hasher.Sum(nil)), nil
}

func encodeResource(typ string, r types.Resource) (*anyv1.Any, error) {
	marshaled, err := proto.MarshalOptions{Deterministic: true}.Marshal(r)
	if err != nil {
		return nil, err
	}

	return &anyv1.Any{TypeUrl: typ, Value: marshaled}, nil
}
