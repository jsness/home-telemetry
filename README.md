# Home Telemetry

## Quickstart (Windows, local server + DB in Docker)

1. Start TimescaleDB and run the server:
```powershell
cd C:\dev\home-telemetry
copy .env.example .env
# edit .env if needed
.\scripts\dev.ps1
```

2. Open Swagger:
```
https://localhost:8443/swagger/index.html
```

## Environment

Required:
- `DATABASE_URL`
- `AUTH_TOKEN`
- `TLS_CERT`
- `TLS_KEY`

Optional:
- `HTTP_ADDR` (default `:8443`)
- `CORS_ORIGINS` (default `*`)

## Migrations

```powershell
cd C:\dev\home-telemetry\server
$env:DATABASE_URL="postgres://postgres:postgres@localhost:5432/telemetry?sslmode=disable"
go run .\cmd\migrate
```

## Notes
- TLS is required; the dev script generates a self-signed cert in `server\certs`.
- The browser will warn about the self-signed cert; accept it for local use.
- Swagger requires `Authorization: Bearer <token>`.
