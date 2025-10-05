package pgconn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conn struct {
	pool      *pgxpool.Pool
	connCfg   *pgxpool.Config
	connected atomic.Bool
}

type Params struct {
	ConnectionString string
}

func New(p Params) (*Conn, error) {
	connCfg, err := pgxpool.ParseConfig(p.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgx connection string: %w", err)
	}
	return &Conn{connCfg: connCfg}, nil
}

func (c *Conn) Run(ctx context.Context) {
	for {
		pool, err := pgxpool.NewWithConfig(ctx, c.connCfg)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			slog.Warn("failed to create db connection pool", slog.String("err", err.Error()))
			continue
		}

		c.pool = pool
		break
	}

	defer func() {
		c.connected.Store(false)
		c.pool.Close()
	}()

	c.connected.Store(true)
	<-ctx.Done()
}

func (c *Conn) Exec(ctx context.Context, sql string, arguments ...any) (CommandTag, error) {
	if !c.connected.Load() {
		return CommandTag{}, ErrConnectionNotReady
	}

	return c.pool.Exec(ctx, sql, arguments...)
}

func (c *Conn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if !c.connected.Load() {
		return &Row{err: ErrConnectionNotReady}
	}

	return c.pool.QueryRow(ctx, sql, args...)
}
