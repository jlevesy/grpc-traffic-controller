package kxds

import "context"

type ServiceChangedEvent struct {
}

type ServiceChangedResult struct {
}

type ServiceUpdateHandler interface {
	OnServiceChanged(ctx context.Context, evt ServiceChangedEvent) (ServiceChangedResult, error)
}
