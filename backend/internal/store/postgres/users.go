package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID    string
	Email string
}

type UsersRepo struct {
	db *pgxpool.Pool
}

func NewUsersRepo(db *pgxpool.Pool) *UsersRepo {
	return &UsersRepo{db: db}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (r *UsersRepo) GetOrCreateByEmail(ctx context.Context, email string) (User, error) {
	email = normalizeEmail(email)

	var u User
	err := r.db.QueryRow(ctx, `SELECT id, email FROM users WHERE email=$1`, email).
		Scan(&u.ID, &u.Email)
	if err == nil {
		return u, nil
	}

	// Важно: users.id генерится в БД (см. миграции).
	err = r.db.QueryRow(ctx, `
		INSERT INTO users(email) VALUES($1)
		ON CONFLICT(email) DO UPDATE SET email=EXCLUDED.email
		RETURNING id, email
	`, email).Scan(&u.ID, &u.Email)

	return u, err
}

func (r *UsersRepo) TouchLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login_at=$1 WHERE id=$2`, time.Now().UTC(), userID)
	return err
}
