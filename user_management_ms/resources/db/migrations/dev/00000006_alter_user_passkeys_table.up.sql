ALTER TABLE user_passkeys
    ADD backup_eligible BIT NOT NULL DEFAULT 0,
        backup_state    BIT NOT NULL DEFAULT 0;
