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

High-level runtime flow:

1. Proxy captures sampled traffic and stores request/response metadata.
2. Replay engine re-executes captured traffic against candidate deployments.
3. Diffing/comparison logic scores drift and breaking changes.
4. Dashboard + API expose results for triage and audit.

Detailed architecture and implementation documents are maintained internally.

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

Update `.env` values for:

- `DIFFSURGE_STORAGE_POSTGRES_URL`
- `DIFFSURGE_STORAGE_REDIS_URL`
- `NEXT_PUBLIC_SUPABASE_URL`
- `NEXT_PUBLIC_SUPABASE_ANON_KEY`
- `SUPABASE_SERVICE_ROLE_KEY`
- `SUPABASE_JWT_SECRET`

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