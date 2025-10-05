package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
)

type store interface {
	CreateList(ctx context.Context, list models.List) (err error)
	GetList(ctx context.Context, listID string) (list models.List, err error)
}

type App struct {
	store store
}

type Params struct {
	Store store
}

func New(p Params) *App {
	return &App{store: p.Store}
}

func (a *App) CreateList(ctx context.Context, name string) (string, error) {
	listID := uuid.NewString()
	if err := a.store.CreateList(ctx, models.List{ID: listID, Name: name}); err != nil {
		return "", fmt.Errorf("create list in db: %w", err)
	}
	return listID, nil
}

func (a *App) GetList(ctx context.Context, listID string) (models.List, error) {
	return a.store.GetList(ctx, listID)
}
