ALTER TABLE journal_columns DROP CONSTRAINT journal_columns_role_check;
ALTER TABLE journal_columns ADD CONSTRAINT journal_columns_role_check
    CHECK (role IN ('trade_result', 'pnl', 'r', 'risk'));
