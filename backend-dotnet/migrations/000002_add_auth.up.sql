-- Add password hash to existing users table
ALTER TABLE users
ADD COLUMN password_hash TEXT;

-- Existing Local User may not have login credentials.
-- Real accounts must have both email and password_hash.
ALTER TABLE users
ADD CONSTRAINT users_auth_credentials_check
CHECK (
    (email IS NULL AND password_hash IS NULL)
    OR
    (email IS NOT NULL AND password_hash IS NOT NULL)
);

-- Make email unique regardless of letter casing
CREATE UNIQUE INDEX idx_users_email_lower_unique
ON users (LOWER(email))
WHERE email IS NOT NULL;

-- Store refresh token sessions
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id UUID NOT NULL
        REFERENCES users(id) ON DELETE CASCADE,

    token_hash TEXT NOT NULL UNIQUE,

    expires_at TIMESTAMPTZ NOT NULL,

    revoked_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id
ON refresh_tokens(user_id);