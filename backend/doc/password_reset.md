# Password reset

## Flow

1. The app submits an email to `POST /api/v1/auth/password-reset/request`.
2. The API responds with the same message whether or not that email belongs to an account.
3. If the account exists, the API emails an 8-digit code. The code expires after 15 minutes and can be tried at most five times.
4. The app submits the email, code, and new password to `POST /api/v1/auth/password-reset/confirm`.
5. The API changes the password, consumes the code, and revokes the user's refresh-token sessions. The user returns to Login.

Codes can be requested once per minute for each account. Only an HMAC hash of a code is stored. In development and testing, if SMTP is not configured, the code is written to the backend console. Production requires SMTP configuration for the reset endpoints to be available.

Apply migration `000005_create_password_reset_codes.up.sql` before using these endpoints.

## SMTP configuration

Set these backend environment variables to send real email:

- `SMTP_HOST` — SMTP server host
- `SMTP_PORT` — SMTP port; defaults to `587` (STARTTLS). Port `465` uses implicit TLS.
- `SMTP_USERNAME` and `SMTP_PASSWORD` — both set for authenticated SMTP, or both omitted
- `SMTP_FROM` — sender email address

The password reset endpoint does not require an authenticated app session. The confirmation endpoint requires a valid email, an 8-digit code, and a password of at least 8 bytes.
