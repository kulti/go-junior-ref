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

type Rows struct {
	err  error
	rows pgx.Rows
}

func (r *Rows) Close() {
	r.rows.Close()
}

func (r *Rows) Err() error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Err()
}

func (r *Rows) CommandTag() pgconn.CommandTag {
	if r.err != nil {
		return pgconn.CommandTag{}
	}
	return r.rows.CommandTag()
}

func (r *Rows) FieldDescriptions() []pgconn.FieldDescription {
	if r.err != nil {
		return nil
	}
	return r.rows.FieldDescriptions()
}

func (r *Rows) Next() bool {
	if r.err != nil {
		return false
	}
	return r.rows.Next()
}

func (r *Rows) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return r.rows.Scan()
}

func (r *Rows) Values() ([]any, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rows.Values()
}

func (r *Rows) RawValues() [][]byte {
	if r.err != nil {
		return nil
	}
	return r.rows.RawValues()
}

func (r *Rows) Conn() *pgx.Conn {
	if r.err != nil {
		return nil
	}
	return r.rows.Conn()
}
