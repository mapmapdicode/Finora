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

## Local PostgreSQL and Redis

Start the local dependencies from the repository root:

```bash
docker compose up -d postgres redis
```

- PostgreSQL is published at `localhost:5432` (`wealthos` database).
- Redis is published at `localhost:6379`.
- Copy `.env.example` to `.env` (or set the same environment variables) before
  running the backend. The application runs migrations automatically when
  `DATABASE_URL` is present.

## Demo auth

Create/login user:
- Register: `POST /api/v1/auth/register`
- Login: `POST /api/v1/auth/login`
- Verify email: `POST /api/v1/auth/verify-email` with `email` and the six-digit `code`
- Resend verification email: `POST /api/v1/auth/resend-verification-email`
- For local quick test, set request header `Authorization: token-dev-token` in `.env`.

New registrations require `confirmPassword` and must verify their email before
they receive a session token. Configure delivery in production with:

```bash
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USERNAME=mailer@example.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=mailer@example.com
SMTP_FROM_NAME=Finora
```

In development, the six-digit code is written to the server log so the mobile
flow can be tested without an SMTP service. Production refuses to start unless
`SMTP_HOST` and `SMTP_FROM` are supplied.

## Notes

- Current implementation uses in-memory store for development bootstrapping.
- Replace with PostgreSQL repositories + SePay/Hermes adapters later.
