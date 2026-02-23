# Driftsurge

**Catch breaking API changes before your users do.**

Driftsurge captures production traffic, replays it against new deployments, and surfaces breaking changes — before a single user is affected. Schema diffing, traffic replay, and drift reports in one platform.

## Architecture

```
┌─────────────┐     ┌──────────┐     ┌──────────────┐
│   Frontend   │────▶│  Proxy   │────▶│  Target API  │
│  (Next.js)   │     │ (Go)     │     └──────────────┘
└──────┬───────┘     └────┬─────┘
       │                  │ captures traffic
       │                  ▼
       │            ┌──────────┐     ┌──────────────┐
       └───────────▶│  API     │────▶│  PostgreSQL   │
                    │  (Go)    │     │  (Supabase)   │
                    └────┬─────┘     └──────────────┘
                         │
                    ┌────▼─────┐     ┌──────────────┐
                    │ Replayer │────▶│    Redis      │
                    │  (Go)    │     │  (Upstash)    │
                    └──────────┘     └──────────────┘
```

| Service | Description | Port |
|---------|-------------|------|
| **Frontend** | Next.js 16 dashboard with Supabase auth | 3000 |
| **API** | Go REST API for projects, traffic, replays, orgs | 8080 |
| **Proxy** | Go reverse proxy that captures API traffic with PII redaction | 8081 |
| **Replayer** | Go worker that replays captured traffic and compares responses | — |

## Tech Stack

- **Frontend**: Next.js 16, React 19, TypeScript, Tailwind CSS, Radix UI, TanStack Query
- **Backend**: Go 1.24, net/http, zerolog, Viper, Prometheus metrics
- **Auth**: Supabase (OAuth + email/password)
- **Database**: PostgreSQL (Supabase hosted)
- **Cache/Queue**: Redis (Upstash hosted)
- **CI**: GitHub Actions (lint, test, build for Go + frontend)

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Node.js 20+ (for local frontend dev)
- Go 1.24+ (for local backend dev)

### Run with Docker Compose

```bash
# Clone the repo
git clone https://github.com/driftsurge/driftsurge.git
cd driftsurge

# Start all services (reads from .env automatically)
docker compose up --build
```

Services will be available at:
- Frontend: http://localhost:3000
- API: http://localhost:8080
- Proxy: http://localhost:8081

### Local Frontend Development

```bash
cd tvc-frontend
npm install
npm run dev
```

### Local Backend Development

```bash
cd tvc-go
go run ./cmd/api     # API server
go run ./cmd/proxy   # Proxy server
go run ./cmd/replayer # Replay worker
```

## Project Structure

```
.
├── docker-compose.yml       # Full-stack orchestration
├── .env                     # Environment config for docker-compose
├── tvc-go/                  # Go backend monorepo
│   ├── cmd/                 # Entry points (api, proxy, replayer, cli)
│   ├── internal/            # Business logic
│   │   ├── api/             # HTTP handlers, middleware, routing
│   │   ├── config/          # Viper config loader
│   │   ├── diffing/         # JSON diff engine
│   │   ├── models/          # Domain models
│   │   ├── pii/             # PII detection & redaction
│   │   ├── proxy/           # Traffic capture, sampling
│   │   ├── replayer/        # Replay engine, comparer
│   │   └── storage/         # Postgres & Redis stores
│   ├── pkg/                 # Shared packages (errors, logger)
│   └── test/                # Integration tests
├── tvc-frontend/            # Next.js frontend
│   ├── app/                 # App Router pages
│   ├── components/          # UI components
│   ├── lib/                 # API client, Supabase, providers
│   └── supabase/migrations/ # Database schema migrations
└── TVC Docs/                # Architecture & development docs
```

## Environment Variables

All configuration is via environment variables. See `.env` (root, for docker-compose) and `tvc-frontend/.env.local` (for Next.js local dev).

| Variable | Used By | Description |
|----------|---------|-------------|
| `TVC_STORAGE_POSTGRES_URL` | Go services | PostgreSQL connection string |
| `TVC_STORAGE_REDIS_URL` | Go services | Redis connection string (rediss:// for TLS) |
| `NEXT_PUBLIC_SUPABASE_URL` | Frontend | Supabase project URL |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | Frontend | Supabase anonymous key |
| `SUPABASE_SERVICE_ROLE_KEY` | Frontend + API | Supabase service role key |
| `SUPABASE_JWT_SECRET` | API | JWT secret for token verification |

## License

All rights reserved.
