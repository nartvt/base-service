-- Migration: Initial database schema
-- Description: Creates users table with all constraints and indices
-- Date: 2025-12-07
-- Author: Base Service Team

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id            BIGSERIAL PRIMARY KEY,
    email         VARCHAR(50) NOT NULL,
    phone_number  VARCHAR(50) NOT NULL,
    username      VARCHAR(50) NOT NULL,
    first_name    VARCHAR(50) NOT NULL,
    last_name     VARCHAR(50) NOT NULL,
    hash_password VARCHAR(255) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Add unique constraints
-- Email must be unique for account recovery
ALTER TABLE users
    ADD CONSTRAINT users_email_key UNIQUE (email);

-- Username must be unique for login
ALTER TABLE users
    ADD CONSTRAINT users_username_key UNIQUE (username);

-- Performance indices for common query patterns

-- Index on created_at for sorting and date range queries
-- DESC order optimizes for "newest first" queries (common pattern)
CREATE INDEX IF NOT EXISTS idx_users_created_at
    ON users(created_at DESC);

-- Index on updated_at for change tracking and sync queries
CREATE INDEX IF NOT EXISTS idx_users_updated_at
    ON users(updated_at DESC);

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts with authentication credentials';
COMMENT ON COLUMN users.id IS 'Auto-incrementing primary key';
COMMENT ON COLUMN users.email IS 'User email address (unique, used for account recovery)';
COMMENT ON COLUMN users.username IS 'User login name (unique, used for authentication)';
COMMENT ON COLUMN users.hash_password IS 'Argon2id hashed password';
COMMENT ON COLUMN users.created_at IS 'Account creation timestamp';
COMMENT ON COLUMN users.updated_at IS 'Last profile update timestamp';

-- Update statistics for query planner
ANALYZE users;
