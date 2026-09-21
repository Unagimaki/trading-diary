ALTER TABLE journal_columns ADD CONSTRAINT journal_columns_id_journal_unique UNIQUE (id, journal_id);

CREATE TABLE journal_rows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id uuid NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (journal_id, position) DEFERRABLE INITIALLY DEFERRED,
    UNIQUE (id, journal_id)
);

CREATE TABLE journal_cells (
    row_id uuid NOT NULL,
    column_id uuid NOT NULL,
    journal_id uuid NOT NULL,
    value jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (row_id, column_id),
    FOREIGN KEY (row_id, journal_id) REFERENCES journal_rows(id, journal_id) ON DELETE CASCADE,
    FOREIGN KEY (column_id, journal_id) REFERENCES journal_columns(id, journal_id) ON DELETE CASCADE
);

CREATE INDEX journal_rows_journal_id_idx ON journal_rows(journal_id);
CREATE INDEX journal_cells_journal_id_idx ON journal_cells(journal_id);

