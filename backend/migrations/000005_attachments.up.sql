CREATE TABLE attachments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    journal_id uuid NOT NULL REFERENCES journals(id) ON DELETE CASCADE,
    row_id uuid NOT NULL,
    column_id uuid NOT NULL,
    storage_key text NOT NULL UNIQUE,
    original_name text NOT NULL,
    mime_type text NOT NULL,
    size_bytes bigint NOT NULL CHECK (size_bytes > 0 AND size_bytes <= 5242880),
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (row_id, column_id),
    FOREIGN KEY (row_id, journal_id) REFERENCES journal_rows(id, journal_id) ON DELETE CASCADE,
    FOREIGN KEY (column_id, journal_id) REFERENCES journal_columns(id, journal_id) ON DELETE CASCADE
);

CREATE INDEX attachments_user_id_idx ON attachments(user_id);

