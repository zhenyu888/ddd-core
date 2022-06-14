package ebus

import (
	"context"
	"sync"
)

type EBus interface {
	Post(ctx context.Context, topic string, event interface{}) error
	Register(handler EventHandler, topic ...string) error
	RegisterAsync(handler EventHandler, topic ...string) error
	Unregister(handler EventHandler, topic ...string)
}

type EventHandler interface {
	Identifier() string
	OnEvent(ctx context.Context, event interface{})
}

type EventFilter interface {
	Filter(event interface{}) bool
}

type bus struct {
	dispatcher Dispatcher
	topics     map[string]*SubscriberRegistry
	lock       sync.RWMutex
}

func NewEBus(opt ...Option) EBus {
	opts := buildOptions(opt...)
	return &bus{
		dispatcher: opts.dispatcher,
		topics:     make(map[string]*SubscriberRegistry),
	}
}

func (b *bus) loadRegistry(topic string) (*SubscriberRegistry, bool) {
	b.lock.RLock()
	b.lock.RUnlock()
	r, ok := b.topics[topic]
	return r, ok
}

func (b *bus) loadOrStoreRegistry(topic string) *SubscriberRegistry {
	if v, ok := b.topics[topic]; ok {
		return v
	}
	b.lock.Lock()
	defer b.lock.Unlock()
	if v, ok := b.topics[topic]; ok {
		return v
	}
	registry := NewSubscriberRegistry()
	b.topics[topic] = registry
	return registry
}

func (b *bus) Post(ctx context.Context, topic string, event interface{}) error {
	if subscribers, ok := b.loadRegistry(topic); ok {
		b.dispatcher.Dispatch(ctx, event, subscribers.GetSubscribers(event))
		return nil
	} else {
		return ErrTopicNotFound
	}
}

func (b *bus) Register(handler EventHandler, topics ...string) error {
	if len(topics) == 0 {
		return ErrRegisterTopicNotSet
	}
	for _, topic := range topics {
		subscribers := b.loadOrStoreRegistry(topic)
		if err := subscribers.Register(handler, false); err != nil {
			return err
		}
	}
	return nil
}

func (b *bus) RegisterAsync(handler EventHandler, topics ...string) error {
	if len(topics) == 0 {
		return ErrRegisterTopicNotSet
	}
	for _, topic := range topics {
		subscribers := b.loadOrStoreRegistry(topic)
		if err := subscribers.Register(handler, true); err != nil {
			return err
		}
	}
	return nil
}

func (b *bus) Unregister(handler EventHandler, topics ...string) {
	if len(topics) == 0 {
		return
	}
	for _, topic := range topics {
		if subscribers, ok := b.loadRegistry(topic); ok {
			subscribers.Unregister(handler)
		}
	}
}

var (
	defaultBus EBus
	once       sync.Once
)

func getDefaultBus() EBus {
	if defaultBus == nil {
		once.Do(func() {
			defaultBus = NewEBus()
		})
	}
	return defaultBus
}

func Post(ctx context.Context, topic string, event interface{}) error {
	return getDefaultBus().Post(ctx, topic, event)
}

func Register(handler EventHandler, topic ...string) error {
	return getDefaultBus().Register(handler, topic...)
}

func RegisterAsync(handler EventHandler, topic ...string) error {
	return getDefaultBus().RegisterAsync(handler, topic...)
}

func Unregister(handler EventHandler, topic ...string) {
	getDefaultBus().Unregister(handler, topic...)
}
