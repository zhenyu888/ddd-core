package ebus

import (
	"container/list"
	"context"
	"fmt"
	"sync"
)

type Subscriber interface {
	Identifier() string
	Filter(event interface{}) bool
	Dispatch(ctx context.Context, event interface{})
}

type SubscriberRegistry struct {
	subscribers *list.List
	idMap       map[string]*list.Element
	lock        sync.RWMutex
}

func NewSubscriberRegistry() *SubscriberRegistry {
	return &SubscriberRegistry{
		subscribers: list.New(),
		idMap:       make(map[string]*list.Element),
	}
}

func (s *SubscriberRegistry) GetSubscribers(event interface{}) []Subscriber {
	s.lock.RLock()
	defer s.lock.RUnlock()

	rlt := make([]Subscriber, 0, s.subscribers.Len())
	ele := s.subscribers.Front()
	for ; ele != nil; ele = ele.Next() {
		sub := ele.Value.(Subscriber)
		if sub.Filter(event) {
			continue
		}
		rlt = append(rlt, sub)
	}
	return rlt
}

func (s *SubscriberRegistry) Register(handler EventHandler, async bool) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := subscriberIdentifier(handler)
	if _, ok := s.idMap[key]; ok {
		return ErrSubscriberAlreadyRegistered
	}
	var subscriber Subscriber
	if async {
		subscriber = newAsyncSubscriber(handler)
	} else {
		subscriber = newSyncSubscriber(handler)
	}
	ele := s.subscribers.PushBack(subscriber)
	s.idMap[key] = ele
	return nil
}

func (s *SubscriberRegistry) Unregister(handler EventHandler) {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := subscriberIdentifier(handler)
	if ele, ok := s.idMap[key]; ok {
		s.subscribers.Remove(ele)
		delete(s.idMap, key)
	}
}

func subscriberIdentifier(handler EventHandler) string {
	return fmt.Sprintf("subscriber-%s", handler.Identifier())
}

type syncSubscriber struct {
	identifier string
	handler    EventHandler
}

func newSyncSubscriber(handler EventHandler) Subscriber {
	return &syncSubscriber{
		identifier: subscriberIdentifier(handler),
		handler:    handler,
	}
}

func (s *syncSubscriber) Identifier() string {
	return s.identifier
}

func (s *syncSubscriber) Filter(event interface{}) bool {
	if f, ok := s.handler.(EventFilter); ok {
		return f.Filter(event)
	}
	return false
}

func (s *syncSubscriber) Dispatch(ctx context.Context, event interface{}) {
	s.handler.OnEvent(ctx, event)
}

type asyncSubscriber struct {
	identifier string
	handler    EventHandler
}

func newAsyncSubscriber(handler EventHandler) Subscriber {
	return &asyncSubscriber{
		identifier: subscriberIdentifier(handler),
		handler:    handler,
	}
}

func (s *asyncSubscriber) Identifier() string {
	return s.identifier
}

func (s *asyncSubscriber) Filter(event interface{}) bool {
	if f, ok := s.handler.(EventFilter); ok {
		return f.Filter(event)
	}
	return false
}

func (s *asyncSubscriber) Dispatch(ctx context.Context, event interface{}) {
	go func() {
		s.handler.OnEvent(ctx, event)
	}()
}
