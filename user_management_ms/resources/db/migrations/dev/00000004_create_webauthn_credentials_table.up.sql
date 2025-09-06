-- -- passkey credentials
CREATE TABLE webauthn_credentials
(
    id            INT IDENTITY (1,1) PRIMARY KEY,
    user_id       INT            NOT NULL FOREIGN KEY REFERENCES users (id) ON DELETE CASCADE,
    credential_id VARBINARY(255) NOT NULL UNIQUE,    -- raw credential ID
    public_key    VARBINARY(255) NOT NULL,           -- COSE public key
    sign_count    BIGINT         NOT NULL DEFAULT 0, -- used to detect clones
    created_at    DATETIME2      NOT NULL DEFAULT GETDATE()
);
