package ebus

import "errors"

var (
	ErrTopicNotFound               = errors.New("topic not found")
	ErrRegisterTopicNotSet         = errors.New("register topic not set")
	ErrSubscriberAlreadyRegistered = errors.New("subscriber already registered")
)
