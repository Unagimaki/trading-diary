UPDATE journal_cells cell
SET value = to_jsonb(j.default_rr), source = 'calculated', updated_at = now()
FROM journal_columns c, journals j
WHERE cell.column_id = c.id
  AND cell.journal_id = c.journal_id
  AND j.id = cell.journal_id
  AND c.role = 'r'
  AND CASE
      WHEN jsonb_typeof(cell.value) = 'number' THEN (cell.value #>> '{}')::numeric <= 0
      ELSE false
  END;
