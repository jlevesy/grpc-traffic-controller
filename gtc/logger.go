package gtc

import (
	"context"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"go.uber.org/zap"
)

type loggerCallbacks struct {
	l *zap.Logger
}

func (l *loggerCallbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	l.l.Debug("New stream opened", zap.Int64("stream_id", id))
	return nil
}

func (l *loggerCallbacks) OnStreamClosed(id int64, n *corev3.Node) {
	l.l.Debug(
		"Stream closed",
		zap.Int64("stream_id", id),
		zap.String("node_id", n.Id),
	)
}

func (l *loggerCallbacks) OnStreamRequest(id int64, req *discoveryv3.DiscoveryRequest) error {
	if req.ErrorDetail != nil {
		l.l.Error(
			"Client NACKed a response",
			zap.String("err", req.ErrorDetail.Message),
			zap.Int64("stream_id", id),
		)
	}

	l.l.Debug(
		"Received a new stream request",
		zap.Strings("resources", req.ResourceNames),
		zap.String("type", req.TypeUrl),
		zap.String("version", req.VersionInfo),
		zap.String("nonce", req.ResponseNonce),
		zap.Int64("stream_id", id),
	)
	return nil
}

func (l *loggerCallbacks) OnStreamResponse(_ context.Context, id int64, req *discoveryv3.DiscoveryRequest, resp *discoveryv3.DiscoveryResponse) {
	l.l.Debug(
		"Sending a new response",
		zap.Int64("stream_id", id),
		zap.String("type", req.TypeUrl),
		zap.Strings("resources", req.ResourceNames),
		zap.String("response_version", resp.VersionInfo),
		zap.String("response_nonce", resp.Nonce),
	)
}
