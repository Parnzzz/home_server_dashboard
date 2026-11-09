ALTER TABLE users
DROP CONSTRAINT users_role_check;

ALTER TABLE users
DROP COLUMN avatar_url,
DROP COLUMN role;