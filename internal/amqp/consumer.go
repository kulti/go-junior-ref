package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	dialAddress     string
	declareExchange Exchange
	declareQueue    Queue
	consumeCh       chan Delivery
	connected       atomic.Bool
}

type ConsumerParams struct {
	DialAddress     string
	DeclareExchange Exchange
	DeclareQueue    Queue
}

type Queue struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Arguments  amqp091.Table
	BindTo     Bind
}

type Bind struct {
	ExchangeName string
	RoutingKey   string
	NoWait       bool
	Arguments    amqp091.Table
}

func NewConsumer(p ConsumerParams) *Consumer {
	return &Consumer{
		dialAddress:     p.DialAddress,
		declareExchange: p.DeclareExchange,
		declareQueue:    p.DeclareQueue,
		consumeCh:       make(chan Delivery),
	}
}

func (c *Consumer) ConsumeCh() <-chan Delivery {
	return c.consumeCh
}

func (c *Consumer) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := c.connectAndConsume(ctx); err != nil {
			slog.Warn("AMQP consumer connection error", slog.String("err", err.Error()))
			time.Sleep(time.Second)
		}
	}
}

func (c *Consumer) Ready() error {
	if !c.connected.Load() {
		return ErrConsumerNotReady
	}
	return nil
}

func (c *Consumer) connectAndConsume(ctx context.Context) error {
	conn, err := amqp091.Dial(c.dialAddress)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}

	if ex := c.declareExchange; ex.Name != "" {
		if err := channel.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDelete, ex.Internal,
			ex.NoWait, ex.Arguments); err != nil {
			return fmt.Errorf("declare exchange: %w", err)
		}
	}

	if q := c.declareQueue; q.Name != "" {
		_, err := channel.QueueDeclare(
			q.Name,
			q.Durable,
			q.AutoDelete,
			q.Exclusive,
			q.NoWait,
			q.Arguments,
		)
		if err != nil {
			return fmt.Errorf("declare queue: %w", err)
		}

		if b := q.BindTo; b.ExchangeName != "" {
			err := channel.QueueBind(
				q.Name,
				b.RoutingKey,
				b.ExchangeName,
				b.NoWait,
				b.Arguments,
			)
			if err != nil {
				return fmt.Errorf("bind queue: %w", err)
			}
		}
	}

	consumeCh, err := channel.ConsumeWithContext(
		ctx,
		c.declareQueue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("start consuming: %w", err)
	}

	c.connected.Store(true)
	defer func() {
		if err := channel.Close(); err != nil {
			slog.Warn("failed to close AMQP channel", slog.String("err", err.Error()))
		}
		c.connected.Store(false)
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case err := <-channel.NotifyClose(make(chan *amqp091.Error, 1)):
			return fmt.Errorf("channel closed: %w", err)
		case msg := <-consumeCh:
			c.consumeCh <- msg
		}
	}
}
