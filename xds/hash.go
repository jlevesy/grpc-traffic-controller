package xds

import corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"

const DefautHashKey = "kxds"

type ConstantHash string

func (h ConstantHash) ID(*corev3.Node) string { return string(h) }

var DefaultHash = ConstantHash(DefautHashKey)
