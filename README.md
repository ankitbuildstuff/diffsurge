# Diffsurge — Traffic Version Control for APIs

![Diffsurge](assets/banner/diffsurge_og.png)

Diffsurge helps teams catch API breaking changes before customers do by combining schema diffing, production traffic capture, and replay validation.

[![Go CI](https://github.com/diffsurge-org/diffsurge/actions/workflows/go.yml/badge.svg)](https://github.com/diffsurge-org/diffsurge/actions/workflows/go.yml)
[![Frontend CI](https://github.com/diffsurge-org/diffsurge/actions/workflows/frontend.yml/badge.svg)](https://github.com/diffsurge-org/diffsurge/actions/workflows/frontend.yml)
[![Release](https://github.com/diffsurge-org/diffsurge/actions/workflows/release.yml/badge.svg)](https://github.com/diffsurge-org/diffsurge/actions/workflows/release.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Why Diffsurge

- **CLI first:** compare schemas and payloads in local dev + CI.
- **Traffic-aware:** capture real API traffic with a low-overhead proxy.
- **Replay validation:** re-run real production requests against new deployments.
- **Governance:** review drift, audit activity, and team-level API changes.

## Architecture at a glance

Diffsurge is a monorepo with three product surfaces:

- **Go services (`diffsurge-go`)**: CLI, API, proxy, and replay engine.
- **Next.js app (`diffsurge-frontend`)**: marketing site + authenticated dashboard.
- **NPM CLI wrapper (`surge-cli-npm`)**: installable distribution of CLI binaries.

```
                        ┌─────────────────────────────────────┐
                        │           End Users / CI             │
                        └───────┬─────────────┬───────────────┘
                                │ surge CLI   │ Browser
                                ▼             ▼
┌───────────────────────────────────────────────────────────────────────────┐
│                               Public Internet                              │
└──────┬──────────────────────────┬────────────────────────┬────────────────┘
       │ :8080                    │ :3000                  │ :8081
       ▼                          ▼                         ▼
┌─────────────┐         ┌──────────────────┐      ┌──────────────────┐
│  API Server │         │ Next.js Frontend │      │  Traffic Proxy   │
│ (diffsurge- │         │ (Dashboard +     │      │ (diffsurge-      │
│    api)     │ ◄──────►│  Marketing site) │      │   proxy)         │
└──────┬──────┘         └──────┬───────────┘      └────────┬─────────┘
       │                       │ Auth (JWT/JWKS)            │ captures
       │ SQL                   │                            │ sampled traffic
       ▼                       ▼                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Supabase (managed cloud)                       │
│  ┌───────────────┐    ┌──────────────┐    ┌───────────────────┐ │
│  │   PostgreSQL  │    │     Auth     │    │  Storage/Realtime │ │
│  │ (traffic +    │    │ (JWT issuer, │    │  (optional)       │ │
│  │  diff data)   │    │  JWKS keys)  │    │                   │ │
│  └───────────────┘    └──────────────┘    └───────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
       ▲
       │ replay results
┌──────┴──────────┐      ┌──────────────┐
│ Replay Engine   │      │  Upstash     │
│ (diffsurge-     │◄────►│  Redis       │
│   replayer)     │      │  (job queue) │
└─────────────────┘      └──────────────┘
```

High-level runtime flow:

1. Proxy captures sampled traffic and stores request/response metadata.
2. Replay engine re-executes captured traffic against candidate deployments.
3. Diffing/comparison logic scores drift and breaking changes.
4. Dashboard + API expose results for triage and audit.

## Quick start

### 1) Prerequisites

- Go 1.24+
- Node.js 20+
- Docker + Docker Compose

### 2) Configure environment

```bash
cp .env.example .env
cp diffsurge-frontend/.env.example diffsurge-frontend/.env.local
```

The table below explains every required variable and which service uses it:

| Variable | Used by | Why it's needed |
|---|---|---|
| `DIFFSURGE_STORAGE_POSTGRES_URL` | api, proxy, replayer | Primary datastore — PostgreSQL connection string (e.g. Supabase DB URL) where traffic captures and diff results are persisted. |
| `DIFFSURGE_STORAGE_REDIS_URL` | api, replayer | Job queue — Upstash/Redis URL used by the replay engine to enqueue and dequeue replay sessions. |
| `NEXT_PUBLIC_SUPABASE_URL` | frontend, api | The Supabase project URL. The frontend uses it to initialize the Supabase JS client for auth flows. The API uses it as the JWKS endpoint (`{url}/auth/v1/jwks`) to verify access tokens without round-tripping Supabase on every request. |
| `NEXT_PUBLIC_SUPABASE_ANON_KEY` | frontend | Public Supabase anon key — safe to expose in the browser. Enables unauthenticated Supabase operations (e.g. sign-up, sign-in) from the Next.js client. |
| `SUPABASE_SERVICE_ROLE_KEY` | frontend (server), api | Supabase service-role key — **secret, never expose client-side**. Used by Next.js API routes to perform privileged DB operations (e.g. user management) and by the Go API as a fallback for admin-level Supabase calls. |
| `SUPABASE_JWT_SECRET` | api | HS256 secret used to verify Supabase-issued JWTs locally. The API validates `Authorization: Bearer <token>` headers against this secret so that authentication does not require a network call to Supabase on every request. |

### 3) Run stack

```bash
docker compose up --build
```

Default local endpoints:

- Frontend: `http://localhost:3000`
- API: `http://localhost:8080`
- Proxy: `http://localhost:8081`

### 4) Run development workflows

```bash
# Go services
cd diffsurge-go
make test
make build

# Frontend
cd ../diffsurge-frontend
npm ci
npm run build
```

## Repository layout

```text
diffsurge/
├── diffsurge-go/          # Go CLI + API + proxy + replay engine
├── diffsurge-frontend/    # Next.js dashboard + marketing site
├── surge-cli-npm/   # NPM packaging for CLI binaries
├── assets/                # Banners and visual media
└── .github/workflows
```

## Open-source roadmap

We are targeting **1,000 GitHub stars in 60 days** after public launch.

- Strategy + milestones: internal planning docs
- Contribution guide: [CONTRIBUTING.md](CONTRIBUTING.md)
- Code of conduct: [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## Contributing

Issues and PRs are welcome. Start with [CONTRIBUTING.md](CONTRIBUTING.md) for setup, quality checks, and PR expectations.

## License

Licensed under the MIT License. See [LICENSE](LICENSE).