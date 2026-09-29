# enmasse-be

Identity and directory API for MASSmaker. Owns users, password hashes, artist/vendor core, claim tokens, and profile completeness.

Massmaker lives in a **separate repo on a separate VPS**. This service does not share Compose, Docker DNS, or Caddy with it.

## Local

```bash
cp .env.example .env
docker compose up -d postgres
go run ./cmd/api
# GET http://localhost:8081/health
```

Seed imported artists (fills `users` + `artists`; `password_hash` stays null):

```bash
go run ./cmd/seed testdata/artist_import.csv
```

Artist signup checks `POST /v1/auth/email/status` first. Seeded emails (`password_hash` null) are `unclaimed` and set a password with `POST /v1/auth/password/claim`. New emails are `available` and use register.

Caddy is optional locally. Massmaker calls this process at `http://127.0.0.1:8081`. The FE does not.

## VPS

```bash
docker compose -f docker-compose.prod.yml up --build -d
```

Caddy serves `api.$ENMASSE_DOMAIN` (health only — auth is 404) and `internal.api.$ENMASSE_DOMAIN` (Massmaker VPS only: `/v1/auth`, `/v1/me`, `/v1/internal`). The Go port and Postgres bind `127.0.0.1`.

## Auth

| Method | Path | Notes |
|---|---|---|
| POST | `/v1/auth/email/status` | `{ status: available \| unclaimed \| registered }` |
| POST | `/v1/auth/register/artist` | New artist (`available` only) |
| POST | `/v1/auth/password/claim` | Set password for seeded `unclaimed` email; returns session |
| POST | `/v1/auth/register/vendor` | Core vendor only — not a public listing |
| POST | `/v1/auth/login` | Body includes `turnstileToken` |
| POST | `/v1/auth/claim` | Always 200 |
| POST | `/v1/auth/password/set` | Token from claim mail |
| GET/PATCH | `/v1/me/profile` | Bearer |
| GET | `/v1/internal/*` | Service JWT from Massmaker |
