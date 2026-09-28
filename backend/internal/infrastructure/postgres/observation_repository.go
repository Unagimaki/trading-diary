package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/observation"
)

type ObservationRepository struct{ pool *pgxpool.Pool }

func NewObservationRepository(pool *pgxpool.Pool) *ObservationRepository {
	return &ObservationRepository{pool: pool}
}
func (r *ObservationRepository) List(ctx context.Context, userID, journalID string) ([]observation.Row, error) {
	rows, err := r.pool.Query(ctx, `SELECT r.id,r.journal_id,r.position,r.created_at,r.updated_at FROM journal_rows r JOIN journals j ON j.id=r.journal_id WHERE r.journal_id=$1 AND j.user_id=$2 ORDER BY r.position`, journalID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]observation.Row, 0)
	byID := map[string]int{}
	for rows.Next() {
		var item observation.Row
		item.Values = map[string]json.RawMessage{}
		item.Warnings = []string{}
		if err := rows.Scan(&item.ID, &item.JournalID, &item.Position, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
		byID[item.ID] = len(items) - 1
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	cells, err := r.pool.Query(ctx, `SELECT c.row_id,c.column_id,c.value FROM journal_cells c JOIN journals j ON j.id=c.journal_id WHERE c.journal_id=$1 AND j.user_id=$2`, journalID, userID)
	if err != nil {
		return nil, err
	}
	defer cells.Close()
	for cells.Next() {
		var rowID, columnID string
		var value json.RawMessage
		if err := cells.Scan(&rowID, &columnID, &value); err != nil {
			return nil, err
		}
		if index, ok := byID[rowID]; ok {
			items[index].Values[columnID] = value
		}
	}
	return items, cells.Err()
}
func (r *ObservationRepository) RoleColumns(ctx context.Context, userID, journalID string) (map[string]string, error) {
	rows, err := r.pool.Query(ctx, `SELECT c.role,c.id FROM journal_columns c JOIN journals j ON j.id=c.journal_id WHERE c.journal_id=$1 AND j.user_id=$2 AND c.role IS NOT NULL ORDER BY c.position`, journalID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	roles := make(map[string]string)
	for rows.Next() {
		var role, columnID string
		if err := rows.Scan(&role, &columnID); err != nil {
			return nil, err
		}
		if _, exists := roles[role]; !exists {
			roles[role] = columnID
		}
	}
	return roles, rows.Err()
}
func (r *ObservationRepository) TradingDefaults(ctx context.Context, userID, journalID string) (observation.TradingDefaults, error) {
	var defaults observation.TradingDefaults
	err := r.pool.QueryRow(ctx, `SELECT initial_deposit,risk_percent,default_rr FROM journals WHERE id=$1 AND user_id=$2`, journalID, userID).Scan(&defaults.Deposit, &defaults.RiskPercent, &defaults.RR)
	if errors.Is(err, pgx.ErrNoRows) {
		return defaults, observation.ErrNotFound
	}
	return defaults, err
}
func (r *ObservationRepository) Create(ctx context.Context, userID, journalID, localDate string) (observation.Row, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return observation.Row{}, err
	}
	defer tx.Rollback(ctx)
	var item observation.Row
	item.Values = map[string]json.RawMessage{}
	item.Warnings = []string{}
	err = tx.QueryRow(ctx, `INSERT INTO journal_rows(journal_id,position) SELECT j.id,COALESCE((SELECT max(position)+1 FROM journal_rows WHERE journal_id=j.id),0) FROM journals j WHERE j.id=$1 AND j.user_id=$2 RETURNING id,journal_id,position,created_at,updated_at`, journalID, userID).Scan(&item.ID, &item.JournalID, &item.Position, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		err = observation.ErrNotFound
	}
	if err != nil {
		return item, err
	}
	rows, err := tx.Query(ctx, `INSERT INTO journal_cells(row_id,column_id,journal_id,value,source) SELECT $1,c.id,$2,to_jsonb($3::text),'default' FROM journal_columns c WHERE c.journal_id=$2 AND c.type='date' RETURNING column_id,value`, item.ID, journalID, localDate)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var columnID string
		var value json.RawMessage
		if err := rows.Scan(&columnID, &value); err != nil {
			rows.Close()
			return item, err
		}
		item.Values[columnID] = value
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return item, err
	}
	rows.Close()
	return item, tx.Commit(ctx)
}
func (r *ObservationRepository) Delete(ctx context.Context, userID, journalID, rowID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var position int
	err = tx.QueryRow(ctx, `DELETE FROM journal_rows r USING journals j WHERE r.id=$1 AND r.journal_id=$2 AND j.id=r.journal_id AND j.user_id=$3 RETURNING r.position`, rowID, journalID, userID).Scan(&position)
	if errors.Is(err, pgx.ErrNoRows) {
		return observation.ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `UPDATE journal_rows SET position=position-1 WHERE journal_id=$1 AND position>$2`, journalID, position); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *ObservationRepository) ColumnType(ctx context.Context, userID, journalID, columnID string) (string, string, []string, error) {
	var typ, role string
	var raw []byte
	err := r.pool.QueryRow(ctx, `SELECT c.type,COALESCE(c.role,''),c.options FROM journal_columns c JOIN journals j ON j.id=c.journal_id WHERE c.id=$1 AND c.journal_id=$2 AND j.user_id=$3`, columnID, journalID, userID).Scan(&typ, &role, &raw)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, observation.ErrNotFound
	}
	var options []string
	if err == nil {
		err = json.Unmarshal(raw, &options)
	}
	return typ, role, options, err
}
func (r *ObservationRepository) SetCell(ctx context.Context, userID, journalID, rowID, columnID string, value json.RawMessage) error {
	tag, err := r.pool.Exec(ctx, `INSERT INTO journal_cells(row_id,column_id,journal_id,value,source) SELECT r.id,c.id,r.journal_id,$5,'manual' FROM journal_rows r JOIN journal_columns c ON c.journal_id=r.journal_id JOIN journals j ON j.id=r.journal_id WHERE r.id=$1 AND c.id=$2 AND r.journal_id=$3 AND j.user_id=$4 ON CONFLICT(row_id,column_id) DO UPDATE SET value=EXCLUDED.value,source='manual',updated_at=now()`, rowID, columnID, journalID, userID, value)
	if err == nil && tag.RowsAffected() == 0 {
		return observation.ErrNotFound
	}
	return err
}
func (r *ObservationRepository) TradingContext(ctx context.Context, userID, journalID, rowID string) (map[string]observation.TradingCell, observation.TradingDefaults, error) {
	var defaults observation.TradingDefaults
	err := r.pool.QueryRow(ctx, `SELECT j.initial_deposit,j.risk_percent,j.default_rr FROM journals j JOIN journal_rows r ON r.journal_id=j.id WHERE j.id=$1 AND j.user_id=$2 AND r.id=$3`, journalID, userID, rowID).Scan(&defaults.Deposit, &defaults.RiskPercent, &defaults.RR)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, defaults, observation.ErrNotFound
	}
	if err != nil {
		return nil, defaults, err
	}
	rows, err := r.pool.Query(ctx, `SELECT c.role,c.id,c.options,cell.value,COALESCE(cell.source,'') FROM journal_columns c JOIN journals j ON j.id=c.journal_id JOIN journal_rows r ON r.journal_id=c.journal_id LEFT JOIN journal_cells cell ON cell.row_id=r.id AND cell.column_id=c.id WHERE c.journal_id=$1 AND j.user_id=$2 AND r.id=$3 AND c.role IS NOT NULL ORDER BY c.position`, journalID, userID, rowID)
	if err != nil {
		return nil, defaults, err
	}
	defer rows.Close()
	result := map[string]observation.TradingCell{}
	for rows.Next() {
		var role, columnID, source string
		var optionsRaw []byte
		var value json.RawMessage
		if err := rows.Scan(&role, &columnID, &optionsRaw, &value, &source); err != nil {
			return nil, defaults, err
		}
		if _, ok := result[role]; ok {
			continue
		}
		var options []string
		if err := json.Unmarshal(optionsRaw, &options); err != nil {
			return nil, defaults, err
		}
		result[role] = observation.TradingCell{ColumnID: columnID, Value: value, Source: source, Options: options}
	}
	return result, defaults, rows.Err()
}
func (r *ObservationRepository) ReplaceCalculated(ctx context.Context, userID, journalID, rowID string, values map[string]json.RawMessage) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `DELETE FROM journal_cells cell USING journal_columns c,journals j WHERE cell.column_id=c.id AND cell.journal_id=$1 AND cell.row_id=$2 AND c.journal_id=cell.journal_id AND c.role IN ('trade_result','pnl','r','risk') AND cell.source='calculated' AND j.id=cell.journal_id AND j.user_id=$3`, journalID, rowID, userID)
	if err != nil {
		return err
	}
	for columnID, value := range values {
		_, err = tx.Exec(ctx, `INSERT INTO journal_cells(row_id,column_id,journal_id,value,source) SELECT r.id,c.id,r.journal_id,$5,'calculated' FROM journal_rows r JOIN journal_columns c ON c.journal_id=r.journal_id JOIN journals j ON j.id=r.journal_id WHERE r.id=$1 AND c.id=$2 AND r.journal_id=$3 AND j.user_id=$4 ON CONFLICT(row_id,column_id) DO UPDATE SET value=EXCLUDED.value,source='calculated',updated_at=now() WHERE journal_cells.source='calculated'`, rowID, columnID, journalID, userID, value)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
func (r *ObservationRepository) ClearCell(ctx context.Context, userID, journalID, rowID, columnID string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM journal_cells cell USING journal_rows r,journal_columns c,journals j WHERE cell.row_id=r.id AND cell.column_id=c.id AND cell.journal_id=$1 AND r.id=$2 AND c.id=$3 AND r.journal_id=cell.journal_id AND c.journal_id=cell.journal_id AND j.id=cell.journal_id AND j.user_id=$4`, journalID, rowID, columnID, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM journal_rows r JOIN journal_columns c ON c.journal_id=r.journal_id JOIN journals j ON j.id=r.journal_id WHERE r.id=$1 AND c.id=$2 AND r.journal_id=$3 AND j.user_id=$4)`, rowID, columnID, journalID, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if !exists {
			return observation.ErrNotFound
		}
	}
	return nil
}
