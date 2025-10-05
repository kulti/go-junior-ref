package pgconn

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type CommandTag = pgconn.CommandTag

type Row struct {
	err error
	row pgx.Row
}

func (r *Row) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.row.Scan(dest...)
}
