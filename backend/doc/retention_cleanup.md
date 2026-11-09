# Database retention cleanup

The API runs retention cleanup at startup and every 24 hours. It deletes in batches of 2,000 rows:

- `metrics` older than 7 days
- revoked `refresh_tokens` older than 30 days, and expired refresh tokens whose expiration is older than 30 days
- `password_reset_codes` created more than 1 day ago

Apply migrations `000005_create_password_reset_codes` and `000006_add_retention_cleanup_indexes` before enabling this version of the API. The indexes support date-based cleanup scans.

`agent_tokens` is bounded to one row per server because token rotation updates the existing row. `users` and `servers` are kept until explicitly deleted; foreign-key cascades remove dependent rows when a server or user is deleted.
