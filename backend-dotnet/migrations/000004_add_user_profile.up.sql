ALTER TABLE users
ADD COLUMN avatar_url TEXT,
ADD COLUMN role VARCHAR(20) NOT NULL DEFAULT 'user';

ALTER TABLE users
ADD CONSTRAINT users_role_check
CHECK (role IN ('admin', 'user', 'viewer'));