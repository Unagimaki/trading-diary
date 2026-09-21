package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/attachment"
)

type AttachmentRepository struct{ pool *pgxpool.Pool }

func NewAttachmentRepository(pool *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{pool: pool}
}
func (r *AttachmentRepository) Upsert(ctx context.Context, userID, journalID, rowID, columnID, key, name, mime string, size int64) (attachment.Attachment, string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return attachment.Attachment{}, "", err
	}
	defer tx.Rollback(ctx)
	var oldKey *string
	err = tx.QueryRow(ctx, `SELECT a.storage_key FROM journal_rows r JOIN journal_columns c ON c.journal_id=r.journal_id JOIN journals j ON j.id=r.journal_id LEFT JOIN attachments a ON a.row_id=r.id AND a.column_id=c.id WHERE r.id=$1 AND c.id=$2 AND r.journal_id=$3 AND j.user_id=$4 AND c.type='image' FOR UPDATE OF r`, rowID, columnID, journalID, userID).Scan(&oldKey)
	if errors.Is(err, pgx.ErrNoRows) {
		return attachment.Attachment{}, "", attachment.ErrNotFound
	}
	if err != nil {
		return attachment.Attachment{}, "", err
	}
	var item attachment.Attachment
	err = tx.QueryRow(ctx, `INSERT INTO attachments(user_id,journal_id,row_id,column_id,storage_key,original_name,mime_type,size_bytes) VALUES($1,$2,$3,$4,$5,$6,$7,$8) ON CONFLICT(row_id,column_id) DO UPDATE SET storage_key=EXCLUDED.storage_key,original_name=EXCLUDED.original_name,mime_type=EXCLUDED.mime_type,size_bytes=EXCLUDED.size_bytes,created_at=now() RETURNING id,original_name,mime_type,size_bytes,created_at`, userID, journalID, rowID, columnID, key, name, mime, size).Scan(&item.ID, &item.Name, &item.MimeType, &item.Size, &item.CreatedAt)
	if err != nil {
		return item, "", err
	}
	value, _ := json.Marshal(item)
	_, err = tx.Exec(ctx, `INSERT INTO journal_cells(row_id,column_id,journal_id,value) VALUES($1,$2,$3,$4) ON CONFLICT(row_id,column_id) DO UPDATE SET value=EXCLUDED.value,updated_at=now()`, rowID, columnID, journalID, value)
	if err != nil {
		return item, "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return item, "", err
	}
	if oldKey != nil {
		return item, *oldKey, nil
	}
	return item, "", nil
}
func (r *AttachmentRepository) Get(ctx context.Context, userID, id string) (attachment.Stored, error) {
	var item attachment.Stored
	err := r.pool.QueryRow(ctx, `SELECT id,original_name,mime_type,size_bytes,created_at,storage_key FROM attachments WHERE id=$1 AND user_id=$2`, id, userID).Scan(&item.ID, &item.Name, &item.MimeType, &item.Size, &item.CreatedAt, &item.StorageKey)
	if errors.Is(err, pgx.ErrNoRows) {
		err = attachment.ErrNotFound
	}
	return item, err
}
func (r *AttachmentRepository) Delete(ctx context.Context, userID, journalID, rowID, columnID string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)
	var key string
	err = tx.QueryRow(ctx, `DELETE FROM attachments a USING journals j WHERE a.row_id=$1 AND a.column_id=$2 AND a.journal_id=$3 AND j.id=a.journal_id AND j.user_id=$4 RETURNING a.storage_key`, rowID, columnID, journalID, userID).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", attachment.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM journal_cells WHERE row_id=$1 AND column_id=$2`, rowID, columnID); err != nil {
		return "", err
	}
	return key, tx.Commit(ctx)
}
