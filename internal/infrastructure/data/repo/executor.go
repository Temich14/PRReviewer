package repo

import (
	"context"
	"database/sql"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func getExecutor(ctx context.Context, fallback Executor) Executor {
	if tx, ok := GetTxFromCtx(ctx); ok {
		return tx
	}
	return fallback
}
