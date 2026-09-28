package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/trade-diary/backend/internal/domain/analytics"
)

type AnalyticsRepository struct{ pool *pgxpool.Pool }

func NewAnalyticsRepository(pool *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{pool: pool}
}

func (r *AnalyticsRepository) Load(ctx context.Context, userID, journalID string) (analytics.Dataset, error) {
	var exists bool
	var observationCount int
	var initialDeposit, riskPercent, defaultRR float64
	err := r.pool.QueryRow(ctx, `
		SELECT
			EXISTS(SELECT 1 FROM journals WHERE id=$1 AND user_id=$2),
			(SELECT count(*) FROM journal_rows r JOIN journals j ON j.id=r.journal_id WHERE r.journal_id=$1 AND j.user_id=$2),
			COALESCE((SELECT initial_deposit FROM journals WHERE id=$1 AND user_id=$2),0),
			COALESCE((SELECT risk_percent FROM journals WHERE id=$1 AND user_id=$2),0),
			COALESCE((SELECT default_rr FROM journals WHERE id=$1 AND user_id=$2),0)
	`, journalID, userID).Scan(&exists, &observationCount, &initialDeposit, &riskPercent, &defaultRR)
	if err != nil {
		return analytics.Dataset{}, err
	}
	if !exists {
		return analytics.Dataset{}, analytics.ErrNotFound
	}

	rows, err := r.pool.Query(ctx, `
		SELECT c.id,c.name,c.type,c.role,c.options
		FROM journal_columns c
		JOIN journals j ON j.id=c.journal_id
		WHERE c.journal_id=$1 AND j.user_id=$2 AND (c.type='select' OR c.role IS NOT NULL)
		ORDER BY c.position
	`, journalID, userID)
	if err != nil {
		return analytics.Dataset{}, err
	}
	defer rows.Close()

	dataset := analytics.Dataset{ObservationCount: observationCount, InitialDeposit: initialDeposit, DefaultRisk: riskPercent, DefaultRR: defaultRR, Columns: make([]analytics.ColumnData, 0), Rows: make([]analytics.RowData, 0)}
	for rows.Next() {
		var column analytics.ColumnData
		var options []byte
		if err := rows.Scan(&column.ID, &column.Name, &column.Type, &column.Role, &options); err != nil {
			return analytics.Dataset{}, err
		}
		if err := json.Unmarshal(options, &column.Options); err != nil {
			return analytics.Dataset{}, err
		}
		dataset.Columns = append(dataset.Columns, column)
	}
	if err := rows.Err(); err != nil {
		return analytics.Dataset{}, err
	}

	journalRows, err := r.pool.Query(ctx, `SELECT r.id FROM journal_rows r JOIN journals j ON j.id=r.journal_id WHERE r.journal_id=$1 AND j.user_id=$2 ORDER BY r.position`, journalID, userID)
	if err != nil {
		return analytics.Dataset{}, err
	}
	rowByID := make(map[string]int)
	for journalRows.Next() {
		var row analytics.RowData
		row.Values = make(map[string]json.RawMessage)
		if err := journalRows.Scan(&row.ID); err != nil {
			journalRows.Close()
			return analytics.Dataset{}, err
		}
		rowByID[row.ID] = len(dataset.Rows)
		dataset.Rows = append(dataset.Rows, row)
	}
	if err := journalRows.Err(); err != nil {
		journalRows.Close()
		return analytics.Dataset{}, err
	}
	journalRows.Close()

	cells, err := r.pool.Query(ctx, `
		SELECT cell.row_id,cell.column_id,cell.value
		FROM journal_cells cell
		JOIN journals j ON j.id=cell.journal_id
		JOIN journal_columns c ON c.id=cell.column_id AND c.journal_id=cell.journal_id
		WHERE cell.journal_id=$1 AND j.user_id=$2 AND (c.type='select' OR c.role IS NOT NULL)
	`, journalID, userID)
	if err != nil {
		return analytics.Dataset{}, err
	}
	defer cells.Close()
	for cells.Next() {
		var rowID, columnID string
		var value json.RawMessage
		if err := cells.Scan(&rowID, &columnID, &value); err != nil {
			return analytics.Dataset{}, err
		}
		if index, ok := rowByID[rowID]; ok {
			dataset.Rows[index].Values[columnID] = value
		}
	}
	return dataset, cells.Err()
}
