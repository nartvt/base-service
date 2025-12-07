-- Rollback: Remove indices from users table
-- Description: Removes performance indices added in migration 001

-- Drop indices
DROP INDEX IF EXISTS idx_users_updated_at;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_phone_number;

-- Note: We don't remove the UNIQUE constraint on username as it may cause data integrity issues
-- If you really need to remove it, uncomment the line below:
-- ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_key;
DROP TABLE IF EXISTS users CASCADE;
