package amqp

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	dialAddress     string
	declareExchange Exchange
	channel         atomic.Pointer[amqp091.Channel]
}

type PublisherParams struct {
	DialAddress     string
	DeclareExchange Exchange
}

type Exchange struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	NoWait     bool
	Arguments  amqp091.Table
}

func NewPublisher(p PublisherParams) *Publisher {
	return &Publisher{
		dialAddress:     p.DialAddress,
		declareExchange: p.DeclareExchange,
	}
}

func (p *Publisher) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if err := p.connectAndWait(ctx); err != nil {
			slog.Warn("AMQP publisher connection error", slog.String("err", err.Error()))
		}
	}
}

func (p *Publisher) Ready() error {
	if p.channel.Load() == nil {
		return ErrPublisherNotReady
	}
	return nil
}

func (p *Publisher) connectAndWait(ctx context.Context) error {
	conn, err := amqp091.Dial(p.dialAddress)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}

	if ex := p.declareExchange; ex.Name != "" {
		if err := channel.ExchangeDeclare(ex.Name, ex.Type, ex.Durable, ex.AutoDelete, ex.Internal,
			ex.NoWait, ex.Arguments); err != nil {
			return fmt.Errorf("declare exchange: %w", err)
		}
	}

	p.channel.Store(channel)
	defer func() {
		if err := channel.Close(); err != nil {
			slog.Warn("failed to close AMQP channel", slog.String("err", err.Error()))
		}
		p.channel.Store(nil)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-channel.NotifyClose(make(chan *amqp091.Error, 1)):
		return fmt.Errorf("channel closed: %w", err)
	}
}

func (p *Publisher) Publish(ctx context.Context, opts PublishOptions) error {
	channel := p.channel.Load()
	if channel == nil {
		return ErrPublisherNotReady
	}

	if err := channel.PublishWithContext(ctx,
		opts.Exchange,
		opts.RoutingKey,
		opts.Mandatory,
		opts.Immediate,
		opts.Msg,
	); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}

	return nil
}
