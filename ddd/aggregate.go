package ddd

import (
	"fmt"
	"time"

	"github.com/zhenyu888/ddd-core/diff"
	"github.com/zhenyu888/ddd-core/funcs"
)

type Aggregate interface {
	AggregateId() int64
}

type AggregateRoot interface {
	Aggregate
	IsZero(aggregate Aggregate) bool
	// RaiseEvent 发布领域事件
	RaiseEvent(event DomainEvent)
	// Events 获取所有的领域事件
	Events() []DomainEvent
	// ClearEvents 清空所有领域事件
	ClearEvents()

	// Attach 对当前聚合根进行快照，开始动态追踪
	Attach(Aggregate)
	// Detach 当初当前聚合根快照，不再追踪
	Detach(aggregateId int64)
	// Snapshot 获取当前聚合根的快照
	Snapshot() Aggregate
	// Diff 当前聚合根跟快照对比改动了哪些属性
	Diff() diff.AggregateDiff
}

type AggregateManager struct {
	events   []DomainEvent
	snapshot Aggregate
}

func (a *AggregateManager) IsZero(aggregate Aggregate) bool {
	return aggregate.AggregateId() <= 0
}

func (a *AggregateManager) RaiseEvent(event DomainEvent) {
	if setter, ok := event.(DomainEventSetter); ok {
		if event.GetEventId() == "" {
			setter.SetEventId(fmt.Sprintf("DomainEvent:%s-%d", funcs.Base64(event.String()), time.Now().UnixNano()))
		}
		if event.GetOccurredOn() <= 0 {
			setter.SetOccurredOn(time.Now().Unix())
		}
	}
	a.events = append(a.events, event)
}

func (a *AggregateManager) Events() []DomainEvent {
	return a.events
}

func (a *AggregateManager) ClearEvents() {
	if len(a.events) == 0 {
		return
	}
	for idx := range a.events {
		a.events[idx] = nil
	}
	a.events = nil
}

func (a *AggregateManager) Attach(aggregate Aggregate) {
	if a.IsZero(aggregate) {
		return
	}
	if a.snapshot == nil || aggregate.AggregateId() == a.snapshot.AggregateId() {
		copied := funcs.DeepCopy(aggregate)
		if agg, ok := copied.(Aggregate); ok {
			a.snapshot = agg
		}
	}
}

func (a *AggregateManager) Detach(aggregateId int64) {
	if a.snapshot != nil && a.snapshot.AggregateId() == aggregateId {
		a.snapshot = nil
	}
}

func (a *AggregateManager) Snapshot() Aggregate {
	return a.snapshot
}

func (a *AggregateManager) Diff() diff.AggregateDiff {
	return nil
}
