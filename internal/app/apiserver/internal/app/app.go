package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/kulti/task_list_course/internal/amqp"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type store interface {
	CreateList(ctx context.Context, list models.List) (err error)
	CreateItem(ctx context.Context, item models.Item) (err error)
	GetList(ctx context.Context, listID string) (list models.List, err error)
	DoneItem(ctx context.Context, itemID string) (item models.Item, err error)
	Subscribe(ctx context.Context, listID, email string) error
}

type publisher interface {
	Publish(ctx context.Context, opts amqp.PublishOptions) error
}

type App struct {
	store     store
	publisher publisher
}

type Params struct {
	Store     store
	Publisher publisher
}

func New(p Params) *App {
	return &App{store: p.Store, publisher: p.Publisher}
}

func (a *App) CreateList(ctx context.Context, name string) (string, error) {
	listID := uuid.NewString()
	if err := a.store.CreateList(ctx, models.List{ID: listID, Name: name}); err != nil {
		return "", fmt.Errorf("create list in db: %w", err)
	}
	return listID, nil
}

func (a *App) CreateItem(ctx context.Context, listID, name string) (string, error) {
	itemID := uuid.NewString()
	if err := a.store.CreateItem(ctx, models.Item{ID: itemID, ListID: listID, Name: name}); err != nil {
		return "", fmt.Errorf("create list in db: %w", err)
	}
	return itemID, nil
}

func (a *App) GetList(ctx context.Context, listID string) (models.List, error) {
	return a.store.GetList(ctx, listID)
}

func (a *App) DoneItem(ctx context.Context, itemID string) (models.Item, error) {
	item, err := a.store.DoneItem(ctx, itemID)

	if err == nil {
		doneEvent := struct {
			Type   string `json:"@type"`
			ItemID string `json:"item_id"`
		}{
			Type:   "item_done",
			ItemID: itemID,
		}
		listEvent := struct {
			ListID string `json:"list_id"`
			Event  any    `json:"event"`
		}{
			ListID: item.ListID,
			Event:  doneEvent,
		}
		msgBody, err := json.Marshal(listEvent)
		if err != nil {
			return item, fmt.Errorf("marshal done item event: %w", err)
		}
		err = a.publisher.Publish(ctx, amqp.PublishOptions{
			Exchange: "item_events",
			Msg: amqp.Publishing{
				Body: msgBody,
			},
		})
		if err != nil {
			return item, fmt.Errorf("publish done item event: %w", err)
		}
	}
	return item, err
}

func (a *App) Subscribe(ctx context.Context, listID, email string) error {
	return a.store.Subscribe(ctx, listID, email)
}
