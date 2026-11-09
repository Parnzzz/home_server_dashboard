DROP TABLE IF EXISTS refresh_tokens;

DROP INDEX IF EXISTS idx_users_email_lower_unique;

ALTER TABLE users
DROP CONSTRAINT IF EXISTS users_auth_credentials_check;

ALTER TABLE users
DROP COLUMN IF EXISTS password_hash;