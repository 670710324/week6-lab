package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

type ReviewAuthor struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"displayName"`
}

type Review struct {
	ID        int64        `json:"id"`
	MemberID  int64        `json:"memberId"`
	MovieID   int64        `json:"movieId"`
	Text      string       `json:"text"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Member    ReviewAuthor `json:"member"`
	Score     *int16       `json:"score"`
}

func (s *Store) ListReviews(ctx context.Context, movieID int64) ([]Review, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT r.id, r.member_id, r.movie_id, r.text, r.created_at, r.updated_at,
			m.display_name, v.score
		FROM reviews r
		JOIN members m ON m.id = r.member_id
		LEFT JOIN votes v ON v.member_id = r.member_id AND v.movie_id = r.movie_id
		WHERE r.movie_id = $1
		ORDER BY r.created_at DESC, r.id DESC`, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	reviews := []Review{}
	for rows.Next() {
		var review Review
		if err := rows.Scan(&review.ID, &review.MemberID, &review.MovieID, &review.Text,
			&review.CreatedAt, &review.UpdatedAt, &review.Member.DisplayName, &review.Score); err != nil {
			return nil, err
		}
		review.Member.ID = review.MemberID
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (s *Store) AddReview(ctx context.Context, memberID, movieID int64, text string) (Review, error) {
	var review Review
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO reviews (member_id, movie_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, member_id, movie_id, text, created_at, updated_at`,
		memberID, movieID, text).Scan(
		&review.ID, &review.MemberID, &review.MovieID, &review.Text,
		&review.CreatedAt, &review.UpdatedAt)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch {
		case pgErr.Code == "23505" && pgErr.ConstraintName == "reviews_member_id_movie_id_key":
			return Review{}, ErrDuplicate
		case pgErr.Code == "23503" && pgErr.ConstraintName == "reviews_movie_id_fkey":
			return Review{}, ErrNoMovie
		}
	}
	return review, err
}

func (s *Store) DeleteReview(ctx context.Context, reviewID, memberID int64, isAdmin bool) error {
	tag, err := s.Pool.Exec(ctx, `
		DELETE FROM reviews
		WHERE id = $1 AND ($3 OR member_id = $2)`, reviewID, memberID, isAdmin)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
