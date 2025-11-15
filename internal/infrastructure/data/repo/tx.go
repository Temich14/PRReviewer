package repo

import (
	"context"
	"database/sql"
)

type SQLTransactor struct {
	db *sql.DB
}

func NewSQLTransactor(db *sql.DB) *SQLTransactor {
	return &SQLTransactor{db: db}
}

func (t *SQLTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)

	if err := fn(ctxWithTx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return rbErr
		}
		return err
	}

	return tx.Commit()
}

type txKey struct{}

func GetTxFromCtx(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)

	return tx, ok
}
