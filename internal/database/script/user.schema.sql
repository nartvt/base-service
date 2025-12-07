CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(255) NOT NULL,
    avatar        VARCHAR(255) NOT NULL,
    phone_number  VARCHAR(20) NOT NULL,
    username      VARCHAR(50) NOT NULL,
    first_name    VARCHAR(50) NOT NULL,
    last_name     VARCHAR(50) NOT NULL,
    hash_password VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ NOT NULL
);
-- Performance indices for common query patterns
-- Index on created_at for sorting and date range queries
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Index on updated_at for change tracking queries
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users(updated_at DESC);

-- Partial index for active user lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Partial index for email and phone number lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Partial index for phone number lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number);
