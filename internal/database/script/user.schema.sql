CREATE TABLE users (
  id            BIGSERIAL PRIMARY KEY,
  email         VARCHAR(50) NOT NULL UNIQUE,
  phone_number  VARCHAR(50) NOT NULL,
  username      VARCHAR(50) NOT NULL UNIQUE,  -- Added UNIQUE constraint (creates index)
  first_name    VARCHAR(50) NOT NULL,
  last_name     VARCHAR(50) NOT NULL,
  hash_password VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ DEFAULT NOW(),
  updated_at    TIMESTAMPTZ DEFAULT NOW()
);

-- Performance indices for common query patterns
-- Index on created_at for sorting and date range queries
CREATE INDEX idx_users_created_at ON users(created_at DESC);

-- Index on updated_at for change tracking queries
CREATE INDEX idx_users_updated_at ON users(updated_at DESC);

-- Partial index for active user lookups (example for future use)
-- CREATE INDEX idx_users_active ON users(username) WHERE deleted_at IS NULL;
