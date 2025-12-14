package amqp

import "github.com/rabbitmq/amqp091-go"

type (
	Publishing = amqp091.Publishing
	Delivery   = amqp091.Delivery
)

type PublishOptions struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
	Msg        Publishing
}

const (
	ExchangeDirect  = amqp091.ExchangeDirect
	ExchangeFanout  = amqp091.ExchangeFanout
	ExchangeTopic   = amqp091.ExchangeTopic
	ExchangeHeaders = amqp091.ExchangeHeaders
)
