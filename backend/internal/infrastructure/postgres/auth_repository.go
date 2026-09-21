package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/auth"
)

type AuthRepository struct{ pool *pgxpool.Pool }

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository { return &AuthRepository{pool: pool} }

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), name text NOT NULL, email text NOT NULL UNIQUE,
 password_hash text NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS sessions (
 token_hash text PRIMARY KEY, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sessions_user_id_idx ON sessions(user_id);`)
	return err
}

func (r *AuthRepository) CreateUser(ctx context.Context, name, email, hash string) (auth.User, error) {
	var u auth.User
	err := r.pool.QueryRow(ctx, `INSERT INTO users(name,email,password_hash) VALUES($1,$2,$3) RETURNING id,name,email`, name, email, hash).Scan(&u.ID, &u.Name, &u.Email)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return u, auth.ErrEmailTaken
	}
	return u, err
}
func (r *AuthRepository) UserByEmail(ctx context.Context, email string) (auth.User, string, error) {
	var u auth.User
	var hash string
	err := r.pool.QueryRow(ctx, `SELECT id,name,email,password_hash FROM users WHERE email=$1`, email).Scan(&u.ID, &u.Name, &u.Email, &hash)
	return u, hash, err
}
func (r *AuthRepository) CreateSession(ctx context.Context, userID, tokenHash string, expires time.Time) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO sessions(token_hash,user_id,expires_at) VALUES($1,$2,$3)`, tokenHash, userID, expires)
	return err
}
func (r *AuthRepository) UserBySession(ctx context.Context, tokenHash string) (auth.User, error) {
	var u auth.User
	err := r.pool.QueryRow(ctx, `SELECT u.id,u.name,u.email FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.token_hash=$1 AND s.expires_at>now()`, tokenHash).Scan(&u.ID, &u.Name, &u.Email)
	if errors.Is(err, pgx.ErrNoRows) {
		return u, auth.ErrInvalidCredentials
	}
	return u, err
}
func (r *AuthRepository) DeleteSession(ctx context.Context, tokenHash string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash=$1`, tokenHash)
	return err
}
