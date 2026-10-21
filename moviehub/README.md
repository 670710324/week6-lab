# MovieHub API

This is the Week 12 Go API alongside the original `kitchen/` Week 6 project.
It uses PostgreSQL and the `movies`, `genres`, and `movie_genres` tables from
the lab. Apply this project's `schema.sql`, then load the course's 100-movie
`seed.sql`. That dataset is not included in this repository; `seed.sample.sql`
provides two movies from the lab examples for a local smoke test.

## Run

Start the repository's PostgreSQL service from the repository root, then create
the MovieHub database once:

```sh
docker compose up -d
psql 'postgres://postgres:secret@localhost:5432/postgres?sslmode=disable' \
  -c 'CREATE DATABASE moviehub'
```

From this directory, apply the schema and sample data, then set the connection
and persistent signing secret:

```sh
export DATABASE_URL='postgres://postgres:secret@localhost:5432/moviehub?sslmode=disable'
psql "$DATABASE_URL" -f schema.sql
psql "$DATABASE_URL" -f seed.sample.sql
openssl rand -hex 32 > .jwt_secret
export JWT_SECRET="$(cat .jwt_secret)"
export ALLOWED_ORIGINS='http://localhost:3000'
go run .
```

Use the course-provided 100-movie `seed.sql` instead of `seed.sample.sql` when
you have that dataset. It is not included in this repository.

Do not commit or regenerate `.jwt_secret` while tickets signed with the existing
secret are still in use. The API listens on `:8080`.

## Routes

| Method | Path | Access |
| --- | --- | --- |
| POST | `/api/auth/register` | Public |
| POST | `/api/auth/login` | Public |
| GET | `/api/movies?q=...` | Public |
| GET | `/api/movies/{id}/reviews` | Public |
| GET | `/api/me` | Bearer ticket |
| POST | `/api/movies/{id}/reviews` | Bearer ticket |
| DELETE | `/api/reviews/{reviewId}` | Owner or admin |
| GET | `/api/admin/members` | Admin |
| GET | `/api/me/wishlist` | Bearer ticket |
| PUT / DELETE | `/api/me/wishlist/{movieId}` | Bearer ticket |
| PUT / DELETE | `/api/movies/{id}/vote` | Bearer ticket |

Admin roles are assigned out of band in the database; registration always
creates a `member`.
