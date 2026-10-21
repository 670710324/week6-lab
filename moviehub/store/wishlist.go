package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

type WishlistItem struct {
	MovieID int64  `json:"movieId"`
	Note    string `json:"note"`
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]WishlistItem, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT movie_id, note
		FROM wishlist_items
		WHERE member_id = $1
		ORDER BY created_at DESC, movie_id`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []WishlistItem{}
	for rows.Next() {
		var item WishlistItem
		if err := rows.Scan(&item.MovieID, &item.Note); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) PutWishlistItem(ctx context.Context, memberID, movieID int64, note string) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (member_id, movie_id, note)
		VALUES ($1, $2, $3)
		ON CONFLICT (member_id, movie_id) DO UPDATE SET note = EXCLUDED.note`,
		memberID, movieID, note)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "wishlist_items_movie_id_fkey" {
		return ErrNoMovie
	}
	return err
}

func (s *Store) DeleteWishlistItem(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		DELETE FROM wishlist_items WHERE member_id = $1 AND movie_id = $2`,
		memberID, movieID)
	return err
}
