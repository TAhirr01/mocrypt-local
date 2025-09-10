CREATE TABLE users
(
    id                    INT IDENTITY (1,1) PRIMARY KEY,
    created_at            DATETIME2     NOT NULL DEFAULT GETDATE(),
    updated_at            DATETIME2              DEFAULT NULL,
    deleted_at            DATETIME2              DEFAULT NULL,
    email                 NVARCHAR(100) NOT NULL UNIQUE,
    phone                 NVARCHAR(100),
    birth_date            DATETIME2,
    password              NVARCHAR(255),
    google_id             NVARCHAR(200),
    --     last_full_auth DATETIME2 NOT NULL DEFAULT GETDATE(),
    --     last_passkey_auth DATETIME2,
    email_verified        BIT                    DEFAULT 0,
    phone_verified        BIT                    DEFAULT 0,
    email_otp_expire_date DATETIME2              DEFAULT NULL,
    phone_otp_expire_date DATETIME2              DEFAULT NULL,
    email_otp             NVARCHAR(100),
    phone_otp             NVARCHAR(100),
    google2_fa_secret     VARCHAR(255),
    is2_fa_verified       BIT                    DEFAULT 0,
    pin_hash              NVARCHAR(255)          DEFAULT NULL,
    --     user_type     NVARCHAR(100) NOT NULL
);