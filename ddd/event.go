package ddd

import (
	"context"
	"fmt"

	"github.com/zhenyu888/ddd-core/ebus"
)

type DomainEvent interface {
	fmt.Stringer
	GetEventId() string
	GetOccurredOn() int64
}

type DomainEventSetter interface {
	SetEventId(string)
	SetOccurredOn(int64)
}

type DomainEventPublisher interface {
	Publish(context.Context, DomainEvent) error
}

var (
	DomainEventPublisherName = "ddd:core:DomainEventPublisher"
	topic                    = "ddd:domain_event_topic"
)

func NewDomainEventPublisher() DomainEventPublisher {
	rlt := LoadOrStoreComponent(&ebusPublisher{}, func() interface{} {
		return &ebusPublisher{}
	})
	return rlt.(DomainEventPublisher)
}

type ebusPublisher struct{}

func (p *ebusPublisher) Name() string {
	return DomainEventPublisherName
}

func (p *ebusPublisher) Publish(ctx context.Context, event DomainEvent) error {
	return ebus.Post(ctx, topic, event)
}

type DomainEventSubscriber func(context.Context, DomainEvent)

type dddEventHandler struct {
	id    string
	fn    DomainEventSubscriber
	async bool
}

func (e *dddEventHandler) Identifier() string {
	return e.id
}

func (e *dddEventHandler) OnEvent(ctx context.Context, event interface{}) {
	if e.async {
		// 异步事件去除事务
		if txCtx, ok := ctx.(*TransactionContext); ok {
			ctx = txCtx.Ctx()
		}
	}
	e.fn(ctx, event.(DomainEvent))
}

func RegisterAsyncEventSubscriber(id string, subscriber DomainEventSubscriber) {
	if err := ebus.RegisterAsync(&dddEventHandler{
		id:    id,
		fn:    subscriber,
		async: true,
	}, topic); err != nil {
		panic(err)
	}
}

func RegisterSyncEventSubscriber(id string, subscriber DomainEventSubscriber) {
	if err := ebus.Register(&dddEventHandler{
		id:    id,
		fn:    subscriber,
		async: false,
	}, topic); err != nil {
		panic(err)
	}
}
