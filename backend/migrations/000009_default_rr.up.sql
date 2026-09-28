ALTER TABLE journals
    ADD COLUMN default_rr numeric(8, 4) NOT NULL DEFAULT 2
    CHECK (default_rr > 0);
