CREATE TABLE journal_columns (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_id uuid NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    name text NOT NULL CHECK (char_length(name) BETWEEN 1 AND 120),
    type text NOT NULL CHECK (type IN ('text', 'number', 'select', 'date', 'boolean', 'image')),
    role text CHECK (role IN ('trade_result', 'pnl', 'r')),
    options jsonb NOT NULL DEFAULT '[]'::jsonb,
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (journal_id, position) DEFERRABLE INITIALLY DEFERRED
);

CREATE INDEX journal_columns_journal_id_idx ON journal_columns(journal_id);
