# Backend - WealthOS API

## Run

```bash
cd backend
go mod download
go run ./cmd/server
```

Default:
- Port: `8080`
- Health: `GET /healthz`
- Ready: `GET /readyz`
- Base API: `/api/v1`

## Demo auth

Create/login user:
- Register: `POST /api/v1/auth/register`
- Login: `POST /api/v1/auth/login`
- For local quick test, set request header `Authorization: token-dev-token` in `.env`.

## Notes

- Current implementation uses in-memory store for development bootstrapping.
- Replace with PostgreSQL repositories + SePay/Hermes adapters later.

