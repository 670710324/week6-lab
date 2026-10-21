package store

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrEmailTaken = errors.New("email taken")
	ErrDuplicate  = errors.New("duplicate")
	ErrNoMovie    = errors.New("movie not found")
)

type Store struct {
	Pool *pgxpool.Pool
}
