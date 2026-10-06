package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txKey struct{}

func QuerierFromContext(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}

	return pool
}

type TxManager struct {
	pool    *pgxpool.Pool
	timeout time.Duration
}

func NewTxManager(pool *pgxpool.Pool, timeout time.Duration) *TxManager {
	return &TxManager{pool: pool, timeout: timeout}
}

const isoLevel = pgx.ReadCommitted

func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return fn(ctx)
	}

	beginCtx, cancel := context.WithTimeout(ctx, m.timeout)
	tx, err := m.pool.BeginTx(beginCtx, pgx.TxOptions{IsoLevel: isoLevel})
	cancel()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = m.rollback(ctx, tx)
			panic(p)
		}
	}()

	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		if rbErr := m.rollback(ctx, tx); rbErr != nil {
			return errors.Join(err, rbErr)
		}
		return err
	}

	commitCtx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	if err := tx.Commit(commitCtx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (m *TxManager) rollback(ctx context.Context, tx pgx.Tx) error {
	// контекст запроса может быть уже отменён, а откатиться и вернуть соединение в пул нужно всё равно
	rbCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), m.timeout)
	defer cancel()

	if err := tx.Rollback(rbCtx); err != nil {
		return fmt.Errorf("rollback transaction: %w", err)
	}

	return nil
}
