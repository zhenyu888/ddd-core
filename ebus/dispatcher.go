package ebus

import "context"

type Dispatcher interface {
	Dispatch(ctx context.Context, event interface{}, subscribers []Subscriber)
}

type immediateDispatcher struct{}

func NewImmediateDispatcher() Dispatcher {
	return &immediateDispatcher{}
}

func (i *immediateDispatcher) Dispatch(ctx context.Context, event interface{}, subscribers []Subscriber) {
	for _, sub := range subscribers {
		sub.Dispatch(ctx, event)
	}
}
