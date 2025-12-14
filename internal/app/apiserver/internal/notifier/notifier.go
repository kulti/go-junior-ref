package notifier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/kulti/task_list_course/internal/amqp"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type store interface {
	AddEvent(ctx context.Context, event models.ListEvent) error
	ProcessEvents(
		ctx context.Context, fn func(ctx context.Context, email string, event models.ListEvent) error,
	) error
}

type consumer interface {
	ConsumeCh() <-chan amqp.Delivery
}

type Params struct {
	Store    store
	Consumer consumer
}

type Notifier struct {
	store    store
	consumer consumer
}

func New(p Params) *Notifier {
	return &Notifier{
		store:    p.Store,
		consumer: p.Consumer,
	}
}

func (n *Notifier) Run(ctx context.Context) {
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		n.consumeEvents(ctx)
	}()
	go func() {
		defer wg.Done()
		n.sendNotifications(ctx)
	}()
	wg.Wait()
}

func (n *Notifier) Ready() error {
	return nil
}

func (n *Notifier) consumeEvents(ctx context.Context) {
	consumeCh := n.consumer.ConsumeCh()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-consumeCh:
			n.processMessage(ctx, msg)
		}
	}
}

func (n *Notifier) processMessage(ctx context.Context, msg amqp.Delivery) {
	var itemEvent struct {
		ListID string          `json:"list_id"`
		Event  json.RawMessage `json:"event"`
	}

	if err := json.Unmarshal(msg.Body, &itemEvent); err != nil {
		slog.Warn("failed to unmarshal item event", slog.String("err", err.Error()))
		msg.Reject(true)
		return
	}

	err := n.store.AddEvent(ctx, models.ListEvent{ListID: itemEvent.ListID, Event: itemEvent.Event})
	if err != nil {
		slog.Warn("failed to add event to store", slog.String("err", err.Error()))
		msg.Reject(true)
		return
	}

	msg.Ack(false)
}

func (n *Notifier) sendNotifications(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		n.processEvents(ctx)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
		}
	}
}

func (n *Notifier) processEvents(ctx context.Context) {
	for {
		if err := n.store.ProcessEvents(ctx, n.sendNotification); err != nil {
			if errors.Is(err, models.ErrNoEvents) {
				return
			}
			slog.Warn("failed to process events", slog.String("err", err.Error()))
		}
	}
}

func (n *Notifier) sendNotification(ctx context.Context, email string, event models.ListEvent) error {
	fmt.Printf("Sending notification to %s about event(list_id: %s, event: %s)\n",
		email, event.ListID, string(event.Event))
	return nil
}
