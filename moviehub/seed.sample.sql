INSERT INTO genres (id, name_en, name_th) VALUES
    (27, 'Horror', 'สยองขวัญ'),
    (28, 'Action', 'บู๊')
ON CONFLICT (id) DO UPDATE
SET name_en = EXCLUDED.name_en, name_th = EXCLUDED.name_th;

INSERT INTO movies (id, title, title_th, release_date, vote_average, poster_path, popularity) VALUES
    (1423191, 'Resident Evil', 'เรสซิเดนท์ อีวิล', '2026-01-01', 7.3,
     '/i7UyjfPio0VFHB9rBUZSFyhOoM8.jpg', 100),
    (460458, 'Resident Evil: Welcome to Raccoon City', 'ผีชีวะ ปฐมบทแห่งเมืองผีดิบ',
     '2022-11-23', 5.9, '/bArhvjRHl535XMaSh9VjInF2mSZ.jpg', 90)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    title_th = EXCLUDED.title_th,
    release_date = EXCLUDED.release_date,
    vote_average = EXCLUDED.vote_average,
    poster_path = EXCLUDED.poster_path,
    popularity = EXCLUDED.popularity;

INSERT INTO movie_genres (movie_id, genre_id, position) VALUES
    (1423191, 27, 0),
    (460458, 28, 0)
ON CONFLICT (movie_id, genre_id) DO UPDATE SET position = EXCLUDED.position;

-- Demo accounts for local smoke tests. These are intentionally lightweight and
-- designed to let the API be exercised without needing a full user registration
-- flow first.
INSERT INTO members (id, email, password_hash, display_name, role) VALUES
    (1, 'admin@moviehub.local', '$2a$10$g965i/snXXw2A09SCzCdH.rLNQYEav9S0IQAnaeiS.mBlplPc80hG', 'MovieHub Admin', 'admin'),
    (2, 'member@moviehub.local', '$2a$10$/YCBolJvm/2CCr6rQ.pK/ehHHwcjz0fKm5rSzq/TXX.nMrRl5UxI6', 'MovieHub Member', 'member')
ON CONFLICT (id) DO UPDATE SET
    email = EXCLUDED.email,
    password_hash = EXCLUDED.password_hash,
    display_name = EXCLUDED.display_name,
    role = EXCLUDED.role;
