# Home Server Dashboard API (.NET)

ASP.NET Core 8 port of the existing Go API. It uses the same PostgreSQL database and migrations, HTTP routes, JSON field names, JWT format, and token tables. Keep the Go service in `backend/`; this is a separate implementation that can run alongside it on a different port.

## Requirements

- .NET 8 SDK
- PostgreSQL with the existing backend migrations applied through `000006`

The migrations are copied into this folder for reference. The ASP.NET service does not apply migrations automatically. Apply them once to the shared database using the existing migration workflow before starting the service. Keep `JWT_SECRET` the same as the Go backend when switching the Flutter app between implementations, so existing access tokens remain valid.

## Configure and run (PowerShell)

Set these environment variables in the terminal that starts the API:

```powershell
$env:DATABASE_URL = "postgres://dashboard_user:your_password@localhost:5432/home_dashboard?sslmode=disable"
$env:JWT_SECRET = "replace-with-at-least-32-random-bytes"
$env:APP_ENV = "development"
$env:PORT = "8081"
dotnet run --project .\backend-dotnet\BackendDotnet.csproj
```

`PORT` defaults to `8080`. Use another port while the Go backend is running. In development/testing without SMTP configuration, password reset codes are written to the API log. To send email, set `SMTP_HOST`, `SMTP_PORT`, `SMTP_FROM`, and optionally both `SMTP_USERNAME` and `SMTP_PASSWORD`. Port 465 uses implicit TLS; other configured ports use STARTTLS.

## Run with Docker Compose

The existing `backend/docker-compose.yml` now includes the ASP.NET API and PostgreSQL. Put `JWT_SECRET` (the same value used by the Go API) in `backend/.env`, or set it in the shell, then run:

```powershell
cd .\backend
docker compose up --build -d
docker compose logs -f dotnet-api
```

The API is available at `http://localhost:8081`. Stop the services with `docker compose down`; this keeps the PostgreSQL data volume. The compose file uses `postgres` as the database hostname from inside Docker.

The API listens on `0.0.0.0`, so a physical phone can connect through the computer's LAN IP when the firewall allows the selected port. Android Emulator uses `10.0.2.2` to reach the host computer.

## Ported routes

- `GET /health`
- `POST /api/v1/auth/register`, `/login`, `/refresh`, `/logout`
- `GET /api/v1/auth/me`
- `POST /api/v1/auth/password-reset/request`, `/confirm`
- `GET /api/v1/servers`, `POST /api/v1/servers`, `GET /api/v1/servers/{id}`, `DELETE /api/v1/servers/{id}`
- `POST` and `DELETE /api/v1/servers/{id}/agent-token`
- `POST /api/v1/metrics` (agent token)
- `GET /api/v1/servers/{id}/metrics` and `/summary`
- `PATCH /api/v1/users/me`

Daily cleanup removes metrics older than 7 days, revoked/expired refresh tokens older than 30 days, and password reset codes older than 1 day. It runs at startup and every 24 hours, in batches of 2,000 rows.
