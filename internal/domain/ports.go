package domain

import (
	"context"
	"time"
)

type Clock interface {
	Now() time.Time
}

type Rand interface {
	Intn(n int) int
}

type Tx interface {
	Exec(ctx context.Context, sql string, arguments ...any) (int64, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type Row interface {
	Scan(dest ...any) error
}

type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, tx Tx) error) error
}

type IDGenerator interface {
	Generate() string
}
