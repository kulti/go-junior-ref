package amqp

import "errors"

var (
	ErrPublisherNotReady = errors.New("AMQP publisher not ready")
	ErrConsumerNotReady  = errors.New("AMQP consumer not ready")
)
