-- -- passkey credentials
CREATE TABLE user_passkeys
(
    id               INT IDENTITY (1,1) PRIMARY KEY,
    user_id          INT            NOT NULL FOREIGN KEY REFERENCES users (id) ON DELETE CASCADE,
    credential_id    VARBINARY(255) NOT NULL UNIQUE,    -- raw credential ID
    public_key       VARBINARY(255) NOT NULL,           -- COSE public key
    sign_count       BIGINT         NOT NULL DEFAULT 0, -- used to detect clones
    created_at       DATETIME2      NOT NULL DEFAULT GETDATE(),
    updated_at       DATETIME2      DEFAULT NULL,
    attestation_type VARCHAR(max),
    aa_guid          VARBINARY(255) NOT NULL,           --device id
    authenticator    NVARCHAR(MAX)
);
