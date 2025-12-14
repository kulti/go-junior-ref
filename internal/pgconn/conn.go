package pgconn

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Conn struct {
	pool          *pgxpool.Pool
	connCfg       *pgxpool.Config
	checkInterval time.Duration
	connected     atomic.Bool
}

type Params struct {
	ConnectionString string
	ChecknInterval   time.Duration
}

func New(p Params) (*Conn, error) {
	connCfg, err := pgxpool.ParseConfig(p.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("parse pgx connection string: %w", err)
	}
	return &Conn{connCfg: connCfg, checkInterval: p.ChecknInterval}, nil
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

	ticker := time.NewTicker(c.checkInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.pool.Ping(ctx); err != nil {
				if c.connected.Load() {
					slog.Warn("connection lost", slog.String("err", err.Error()))
					c.connected.Store(false)
				}
			} else {
				if !c.connected.Load() {
					slog.Info("connection restored")
					c.connected.Store(true)
				}
			}
		}
	}
}

func (c *Conn) Ready() error {
	if !c.connected.Load() {
		return ErrConnectionNotReady
	}
	return nil
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

func (c *Conn) Query(ctx context.Context, sql string, args ...any) pgx.Rows {
	if !c.connected.Load() {
		return &Rows{err: ErrConnectionNotReady}
	}

	rows, _ := c.pool.Query(ctx, sql, args...)
	return rows
}

func (c *Conn) Begin(ctx context.Context) (pgx.Tx, error) {
	if !c.connected.Load() {
		return nil, ErrConnectionNotReady
	}

	return c.pool.Begin(ctx)
}
