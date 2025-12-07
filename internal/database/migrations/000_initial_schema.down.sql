-- Rollback: Drop initial schema
-- Description: Removes users table and all associated objects
-- WARNING: This will delete all user data!

-- Drop indices first (faster than dropping with table)
DROP INDEX IF EXISTS idx_users_updated_at;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_phone_number;

-- Drop constraints (will be dropped with table, but explicit is better)
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_username_key;
ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_email_key;

-- Drop table (CASCADE will drop dependent objects)
DROP TABLE IF EXISTS users CASCADE;
