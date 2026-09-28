INSERT INTO journal_cells(row_id, column_id, journal_id, value, source)
SELECT r.id, c.id, r.journal_id, to_jsonb(j.risk_percent), 'calculated'
FROM journal_rows r
JOIN journals j ON j.id = r.journal_id
JOIN journal_columns c ON c.journal_id = r.journal_id AND c.role = 'risk'
ON CONFLICT(row_id, column_id) DO NOTHING;

INSERT INTO journal_cells(row_id, column_id, journal_id, value, source)
SELECT r.id, c.id, r.journal_id, to_jsonb(j.default_rr), 'calculated'
FROM journal_rows r
JOIN journals j ON j.id = r.journal_id
JOIN journal_columns c ON c.journal_id = r.journal_id AND c.role = 'r'
ON CONFLICT(row_id, column_id) DO NOTHING;

WITH trade_values AS (
    SELECT DISTINCT ON (r.id, pnl_column.id)
        r.id AS row_id,
        r.journal_id,
        pnl_column.id AS pnl_column_id,
        lower(result_cell.value #>> '{}') AS result,
        j.initial_deposit,
        COALESCE((risk_cell.value #>> '{}')::numeric, j.risk_percent) AS risk_percent,
        COALESCE((rr_cell.value #>> '{}')::numeric, j.default_rr) AS rr
    FROM journal_rows r
    JOIN journals j ON j.id = r.journal_id
    JOIN journal_columns result_column ON result_column.journal_id = r.journal_id AND result_column.role = 'trade_result'
    JOIN journal_cells result_cell ON result_cell.row_id = r.id AND result_cell.column_id = result_column.id
    JOIN journal_columns pnl_column ON pnl_column.journal_id = r.journal_id AND pnl_column.role = 'pnl'
    LEFT JOIN journal_columns risk_column ON risk_column.journal_id = r.journal_id AND risk_column.role = 'risk'
    LEFT JOIN journal_cells risk_cell ON risk_cell.row_id = r.id AND risk_cell.column_id = risk_column.id
    LEFT JOIN journal_columns rr_column ON rr_column.journal_id = r.journal_id AND rr_column.role = 'r'
    LEFT JOIN journal_cells rr_cell ON rr_cell.row_id = r.id AND rr_cell.column_id = rr_column.id
    ORDER BY r.id, pnl_column.id, result_column.position, risk_column.position, rr_column.position
)
INSERT INTO journal_cells(row_id, column_id, journal_id, value, source)
SELECT
    row_id,
    pnl_column_id,
    journal_id,
    to_jsonb(CASE result
        WHEN 'win' THEN initial_deposit * risk_percent / 100 * rr
        WHEN 'loss' THEN -initial_deposit * risk_percent / 100
        WHEN 'breakeven' THEN 0
    END),
    'calculated'
FROM trade_values
WHERE result IN ('win', 'loss', 'breakeven') AND risk_percent > 0 AND risk_percent <= 100 AND rr > 0
ON CONFLICT(row_id, column_id) DO UPDATE
SET value = EXCLUDED.value, source = 'calculated', updated_at = now()
WHERE journal_cells.source = 'calculated';
