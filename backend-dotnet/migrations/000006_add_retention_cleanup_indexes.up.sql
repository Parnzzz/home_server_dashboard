CREATE INDEX idx_metrics_recorded_at
ON metrics(recorded_at);

CREATE INDEX idx_refresh_tokens_expires_at
ON refresh_tokens(expires_at);

CREATE INDEX idx_refresh_tokens_revoked_at
ON refresh_tokens(revoked_at)
WHERE revoked_at IS NOT NULL;

CREATE INDEX idx_password_reset_codes_created_at
ON password_reset_codes(created_at);
