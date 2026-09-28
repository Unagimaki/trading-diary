ALTER TABLE journals
    ADD COLUMN initial_deposit numeric(20, 2) NOT NULL DEFAULT 10000 CHECK (initial_deposit > 0),
    ADD COLUMN risk_percent numeric(8, 4) NOT NULL DEFAULT 1 CHECK (risk_percent > 0 AND risk_percent <= 100);
