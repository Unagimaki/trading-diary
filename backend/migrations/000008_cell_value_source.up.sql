ALTER TABLE journal_cells
    ADD COLUMN source text NOT NULL DEFAULT 'manual'
    CHECK (source IN ('manual', 'calculated', 'default'));
