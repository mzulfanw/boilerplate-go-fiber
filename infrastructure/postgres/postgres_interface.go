package postgres

import "github.com/jackc/pgx/v5/pgxpool"

type DB interface {
	Pool() *pgxpool.Pool
	Close()
}
