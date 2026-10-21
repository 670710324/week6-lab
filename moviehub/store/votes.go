package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Store) PutVote(ctx context.Context, memberID, movieID int64, score int16) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO votes (member_id, movie_id, score)
		VALUES ($1, $2, $3)
		ON CONFLICT (member_id, movie_id) DO UPDATE SET score = EXCLUDED.score`,
		memberID, movieID, score)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "votes_movie_id_fkey" {
		return ErrNoMovie
	}
	return err
}

func (s *Store) DeleteVote(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM votes WHERE member_id = $1 AND movie_id = $2`, memberID, movieID)
	return err
}
