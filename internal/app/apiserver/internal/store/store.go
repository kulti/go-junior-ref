package store

import (
	"context"
	"fmt"

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

func (s *Store) GetList(ctx context.Context, listID string) (models.List, error) {
	list := models.List{ID: listID}
	row := s.conn.QueryRow(ctx, `SELECT name FROM lists WHERE id = $1`, listID)
	if err := row.Scan(&list.Name); err != nil {
		return models.List{}, fmt.Errorf("scan list row: %w", err)
	}
	return list, nil
}
