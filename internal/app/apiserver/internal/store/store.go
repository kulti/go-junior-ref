package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/kulti/task_list_course/internal/app/apiserver/internal/models"
	"github.com/kulti/task_list_course/internal/pgconn"
)

type Store struct {
	conn *pgconn.Conn
}

type Params struct {
	PgConn *pgconn.Conn
}

func New(p Params) *Store {
	return &Store{conn: p.PgConn}
}

func (s *Store) CreateList(ctx context.Context, list models.List) error {
	if _, err := s.conn.Exec(ctx, `INSERT INTO lists(id, name) VALUES($1, $2)`, list.ID, list.Name); err != nil {
		return fmt.Errorf("insert list info: %w", err)
	}
	return nil
}

func (s *Store) CreateItem(ctx context.Context, item models.Item) error {
	if _, err := s.conn.Exec(ctx, `INSERT INTO items(id, list_id, name) VALUES($1, $2, $3)`,
		item.ID, item.ListID, item.Name); err != nil {
		return fmt.Errorf("insert item info: %w", err)
	}
	return nil
}

func (s *Store) GetList(ctx context.Context, listID string) (models.List, error) {
	list := models.List{ID: listID}
	row := s.conn.QueryRow(ctx, `SELECT name FROM lists WHERE id = $1`, listID)
	if err := row.Scan(&list.Name); err != nil {
		return models.List{}, fmt.Errorf("scan list: %w", err)
	}

	rows := s.conn.Query(ctx, `SELECT id, name, done FROM items WHERE list_id = $1`, listID)
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (models.Item, error) {
		var item models.Item
		err := row.Scan(&item.ID, &item.Name, &item.Done)
		return item, err
	})
	if err != nil {
		return models.List{}, fmt.Errorf("scan list items: %w", err)
	}

	list.Items = items
	return list, nil
}

func (s *Store) DoneItem(ctx context.Context, itemID string) (models.Item, error) {
	var item models.Item
	row := s.conn.QueryRow(ctx,
		`UPDATE items SET done = TRUE WHERE id = $1 RETURNING id, list_id, name, done`,
		itemID)
	if err := row.Scan(&item.ID, &item.ListID, &item.Name, &item.Done); err != nil {
		return models.Item{}, fmt.Errorf("scan done item: %w", err)
	}
	return item, nil
}
