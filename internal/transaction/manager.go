package transaction

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/ptalbayeva/go-musthave-diploma/internal/repository"
)

type manager struct {
	db *sqlx.DB
}

func NewTxManager(db *sqlx.DB) repository.TxManager {
	return &manager{db: db}
}

func (m *manager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
