package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/journal"
)

type JournalRepository struct{ pool *pgxpool.Pool }

func NewJournalRepository(pool *pgxpool.Pool) *JournalRepository {
	return &JournalRepository{pool: pool}
}
func scanJournal(row pgx.Row) (journal.Journal, error) {
	var j journal.Journal
	err := row.Scan(&j.ID, &j.Name, &j.CreatedAt, &j.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = journal.ErrNotFound
	}
	return j, err
}
func (r *JournalRepository) List(ctx context.Context, userID string) ([]journal.Journal, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,created_at,updated_at FROM journals WHERE user_id=$1 ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]journal.Journal, 0)
	for rows.Next() {
		j, err := scanJournal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, j)
	}
	return result, rows.Err()
}
func (r *JournalRepository) Create(ctx context.Context, userID, name string) (journal.Journal, error) {
	return scanJournal(r.pool.QueryRow(ctx, `INSERT INTO journals(user_id,name) VALUES($1,$2) RETURNING id,name,created_at,updated_at`, userID, name))
}
func (r *JournalRepository) Get(ctx context.Context, userID, id string) (journal.Journal, error) {
	return scanJournal(r.pool.QueryRow(ctx, `SELECT id,name,created_at,updated_at FROM journals WHERE id=$1 AND user_id=$2`, id, userID))
}
func (r *JournalRepository) Rename(ctx context.Context, userID, id, name string) (journal.Journal, error) {
	return scanJournal(r.pool.QueryRow(ctx, `UPDATE journals SET name=$1,updated_at=now() WHERE id=$2 AND user_id=$3 RETURNING id,name,created_at,updated_at`, name, id, userID))
}
func (r *JournalRepository) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM journals WHERE id=$1 AND user_id=$2`, id, userID)
	if err == nil && tag.RowsAffected() == 0 {
		return journal.ErrNotFound
	}
	return err
}
