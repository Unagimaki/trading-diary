DROP TABLE IF EXISTS journal_cells;
DROP TABLE IF EXISTS journal_rows;
ALTER TABLE journal_columns DROP CONSTRAINT IF EXISTS journal_columns_id_journal_unique;

