package store

import (
	"context"
)

type Movie struct {
	ID      int64   `json:"id"`
	Title   string  `json:"title"`
	TitleTh *string `json:"titleTh"`
	Genre   *string `json:"genre"`
	Year    *int    `json:"year"`
	Rating  float64 `json:"rating"`
	Poster  *string `json:"poster"`
}

func (s *Store) SearchMovies(ctx context.Context, query string) ([]Movie, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT m.id, m.title, m.title_th,
			(SELECT coalesce(g.name_th, g.name_en)
			 FROM movie_genres mg
			 JOIN genres g ON g.id = mg.genre_id
			 WHERE mg.movie_id = m.id AND mg.position = 0),
			extract(year FROM m.release_date)::int,
			round(m.vote_average, 1)::float8,
			CASE WHEN m.poster_path IS NULL THEN NULL
				ELSE 'https://image.tmdb.org/t/p/w500' || m.poster_path END
		FROM movies m
		WHERE lower(m.title) LIKE '%' || lower($1) || '%'
		ORDER BY m.popularity DESC, m.id`, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	movies := []Movie{}
	for rows.Next() {
		var movie Movie
		if err := rows.Scan(&movie.ID, &movie.Title, &movie.TitleTh, &movie.Genre,
			&movie.Year, &movie.Rating, &movie.Poster); err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, rows.Err()
}
