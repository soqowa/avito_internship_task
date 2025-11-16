package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	domain "github.com/user/reviewer-svc/internal/domain"
)


type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Pool() *pgxpool.Pool {
	return m.pool
}

func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context, t domain.Tx) error) error {
	pgxTx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	wrapped := &txWrapper{Tx: pgxTx}

	if err := fn(ctx, wrapped); err != nil {
		if rbErr := pgxTx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx error: %w, rollback error: %v", err, rbErr)
		}
		return err
	}

	return pgxTx.Commit(ctx)
}

type txWrapper struct {
	pgx.Tx
}

func (t *txWrapper) Exec(ctx context.Context, sql string, args ...any) (int64, error) {
	cmd, err := t.Tx.Exec(ctx, sql, args...)
	return cmd.RowsAffected(), err
}

func (t *txWrapper) Query(ctx context.Context, sql string, args ...any) (domain.Rows, error) {
	rows, err := t.Tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (t *txWrapper) QueryRow(ctx context.Context, sql string, args ...any) domain.Row {
	return t.Tx.QueryRow(ctx, sql, args...)
}

var _ domain.TxManager = (*TxManager)(nil)
