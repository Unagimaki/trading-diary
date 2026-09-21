package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/column"
)

type ColumnRepository struct{ pool *pgxpool.Pool }

func NewColumnRepository(pool *pgxpool.Pool) *ColumnRepository { return &ColumnRepository{pool: pool} }
func scanColumn(row pgx.Row) (column.Column, error) {
	var c column.Column
	var options []byte
	err := row.Scan(&c.ID, &c.JournalID, &c.Name, &c.Type, &c.Role, &options, &c.Position, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, column.ErrNotFound
	}
	if err == nil {
		err = json.Unmarshal(options, &c.Options)
	}
	return c, err
}

const columnFields = `c.id,c.journal_id,c.name,c.type,c.role,c.options,c.position,c.created_at,c.updated_at`

func (r *ColumnRepository) List(ctx context.Context, userID, journalID string) ([]column.Column, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+columnFields+` FROM journal_columns c JOIN journals j ON j.id=c.journal_id WHERE c.journal_id=$1 AND j.user_id=$2 ORDER BY c.position`, journalID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]column.Column, 0)
	for rows.Next() {
		c, err := scanColumn(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, c)
	}
	return items, rows.Err()
}
func (r *ColumnRepository) Get(ctx context.Context, userID, journalID, id string) (column.Column, error) {
	return scanColumn(r.pool.QueryRow(ctx, `SELECT `+columnFields+` FROM journal_columns c JOIN journals j ON j.id=c.journal_id WHERE c.id=$1 AND c.journal_id=$2 AND j.user_id=$3`, id, journalID, userID))
}
func (r *ColumnRepository) Create(ctx context.Context, userID, journalID string, v column.Values) (column.Column, error) {
	options, _ := json.Marshal(v.Options)
	return scanColumn(r.pool.QueryRow(ctx, `INSERT INTO journal_columns(journal_id,name,type,role,options,position) SELECT j.id,$3,$4,$5,$6,COALESCE((SELECT max(position)+1 FROM journal_columns WHERE journal_id=j.id),0) FROM journals j WHERE j.id=$1 AND j.user_id=$2 RETURNING id,journal_id,name,type,role,options,position,created_at,updated_at`, journalID, userID, v.Name, v.Type, v.Role, options))
}
func (r *ColumnRepository) Update(ctx context.Context, userID, journalID, id string, v column.Values) (column.Column, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return column.Column{}, err
	}
	defer tx.Rollback(ctx)
	var old, count int
	err = tx.QueryRow(ctx, `SELECT c.position,(SELECT count(*) FROM journal_columns WHERE journal_id=c.journal_id) FROM journal_columns c JOIN journals j ON j.id=c.journal_id WHERE c.id=$1 AND c.journal_id=$2 AND j.user_id=$3 FOR UPDATE`, id, journalID, userID).Scan(&old, &count)
	if errors.Is(err, pgx.ErrNoRows) {
		return column.Column{}, column.ErrNotFound
	}
	if err != nil {
		return column.Column{}, err
	}
	position := old
	if v.Position != nil {
		position = *v.Position
		if position < 0 {
			position = 0
		}
		if position >= count {
			position = count - 1
		}
		if position < old {
			_, err = tx.Exec(ctx, `UPDATE journal_columns SET position=position+1 WHERE journal_id=$1 AND position >= $2 AND position < $3`, journalID, position, old)
		} else if position > old {
			_, err = tx.Exec(ctx, `UPDATE journal_columns SET position=position-1 WHERE journal_id=$1 AND position > $2 AND position <= $3`, journalID, old, position)
		}
		if err != nil {
			return column.Column{}, err
		}
	}
	options, _ := json.Marshal(v.Options)
	c, err := scanColumn(tx.QueryRow(ctx, `UPDATE journal_columns c SET name=$1,type=$2,role=$3,options=$4,position=$5,updated_at=now() FROM journals j WHERE c.id=$6 AND c.journal_id=$7 AND j.id=c.journal_id AND j.user_id=$8 RETURNING c.id,c.journal_id,c.name,c.type,c.role,c.options,c.position,c.created_at,c.updated_at`, v.Name, v.Type, v.Role, options, position, id, journalID, userID))
	if err != nil {
		return c, err
	}
	return c, tx.Commit(ctx)
}
func (r *ColumnRepository) Delete(ctx context.Context, userID, journalID, id string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var position int
	err = tx.QueryRow(ctx, `DELETE FROM journal_columns c USING journals j WHERE c.id=$1 AND c.journal_id=$2 AND j.id=c.journal_id AND j.user_id=$3 RETURNING c.position`, id, journalID, userID).Scan(&position)
	if errors.Is(err, pgx.ErrNoRows) {
		return column.ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE journal_columns SET position=position-1 WHERE journal_id=$1 AND position>$2`, journalID, position); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
