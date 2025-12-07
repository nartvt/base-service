-- Migration: Add performance indices to users table
-- Description: Adds indices for commonly queried columns to improve performance
-- Date: 2025-12-07

-- Add UNIQUE constraint to username (if not exists)
-- This creates an index automatically and enforces uniqueness
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'users_username_key'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
    END IF;
END $$;

-- Index on created_at for sorting and date range queries
-- DESC order optimizes for "newest first" queries (common pattern)
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at DESC);

-- Index on updated_at for change tracking and sync queries
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON users(updated_at DESC);

-- Partial index for active user lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username) WHERE deleted_at IS NULL;

-- Partial index for email and phone number lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE deleted_at IS NULL;

-- Partial index for phone number lookups (example for future use)
CREATE INDEX IF NOT EXISTS idx_users_phone_number ON users(phone_number) WHERE deleted_at IS NULL;

-- Analyze table to update query planner statistics
ANALYZE users;
