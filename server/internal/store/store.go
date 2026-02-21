package store

import "github.com/jackc/pgx/v5/pgxpool"

type Stores struct {
	Pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) Stores {
	return Stores{Pool: pool}
}
