package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
)

type Member struct {
	ID           int64      `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"displayName"`
	Role         string     `json:"role"`
	CreatedAt    *time.Time `json:"createdAt,omitempty"` // มีเฉพาะตอน list ให้ admin ดู
	PasswordHash string     `json:"-"`                   // ไม่มีวันออกไปกับ JSON
}

func (s *Store) CreateMember(ctx context.Context, email, passwordHash, displayName string) (Member, error) {
	member := Member{Email: email, DisplayName: displayName}
	err := s.Pool.QueryRow(ctx, `
		INSERT INTO members (email, password_hash, display_name)
		VALUES ($1, $2, $3)
		RETURNING id, role`, email, passwordHash, displayName).Scan(&member.ID, &member.Role)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return Member{}, ErrEmailTaken
	}
	return member, err
}

func (s *Store) MemberByEmail(ctx context.Context, email string) (Member, error) {
	var member Member
	err := s.Pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, password_hash
		FROM members WHERE lower(email) = $1`, email).
		Scan(&member.ID, &member.Email, &member.DisplayName, &member.Role, &member.PasswordHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return member, err
}

func (s *Store) MemberByID(ctx context.Context, id int64) (Member, error) {
	var member Member
	err := s.Pool.QueryRow(ctx, `
		SELECT id, email, display_name, role, created_at
		FROM members WHERE id = $1`, id).
		Scan(&member.ID, &member.Email, &member.DisplayName, &member.Role, &member.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return member, err
}

func (s *Store) ListMembers(ctx context.Context) ([]Member, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, email, display_name, role, created_at
		FROM members ORDER BY created_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	members := []Member{}
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.ID, &member.Email, &member.DisplayName, &member.Role, &member.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, rows.Err()
}

// UpdateMemberRole เปลี่ยนบทบาท (member <-> admin) ไม่พบคนตอบ ErrNotFound
func (s *Store) UpdateMemberRole(ctx context.Context, id int64, role string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx,
		`UPDATE members SET role = $2, updated_at = now()
         WHERE id = $1
         RETURNING id, email, display_name, role, created_at`, id, role).
		Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role, &m.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return m, err
}
