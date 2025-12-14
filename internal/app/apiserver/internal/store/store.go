package store

import (
	"context"
	"errors"
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

func (s *Store) Subscribe(ctx context.Context, listID, email string) error {
	_, err := s.conn.Exec(ctx, `
		WITH seq_num AS (
			SELECT COALESCE(MAX(seq_num), 0) AS max
			FROM list_events
			WHERE list_id = $1
		)
		INSERT INTO list_subscribers(list_id, email, processed_count)
		VALUES($1, $2, (SELECT max FROM seq_num))
	`, listID, email)
	if err != nil {
		return fmt.Errorf("insert list subscriber: %w", err)
	}
	return nil
}

func (s *Store) AddEvent(ctx context.Context, event models.ListEvent) error {
	_, err := s.conn.Exec(ctx,
		`INSERT INTO list_events(list_id, event_data) VALUES($1, $2)`,
		event.ListID, event.Event)
	if err != nil {
		return fmt.Errorf("insert list event: %w", err)
	}
	return nil
}

func (s *Store) ProcessEvents(
	ctx context.Context,
	fn func(ctx context.Context, email string, event models.ListEvent) error,
) error {
	return pgx.BeginFunc(ctx, s.conn, func(tx pgx.Tx) error {
		var email string
		var event models.ListEvent
		err := tx.QueryRow(ctx, `
			SELECT ls.email, le.list_id, le.event_data
			FROM list_subscribers AS ls
			JOIN list_events as le
				ON ls.list_id = le.list_id
				AND ls.processed_count + 1 = le.seq_num
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`).Scan(&email, &event.ListID, &event.Event)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return models.ErrNoEvents
			}
			return fmt.Errorf("scan event: %w", err)
		}
		fmt.Println("Processing event for ", email, " event: ", string(event.Event))
		if err := fn(ctx, email, event); err != nil {
			return fmt.Errorf("process event: %w", err)
		}

		_, err = tx.Exec(ctx, `
			UPDATE list_subscribers
			SET processed_count = processed_count + 1
			WHERE list_id = $1 AND email = $2
		`, event.ListID, email)
		if err != nil {
			return fmt.Errorf("update processed count: %w", err)
		}
		return nil
	})
}
