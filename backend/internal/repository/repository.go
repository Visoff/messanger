package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (r *Queries) NewTx(ctx context.Context) (*Queries, pgx.Tx, error) {
	tx, err := r.db.(*pgxpool.Pool).Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	return &Queries{db: tx}, tx, nil
}
