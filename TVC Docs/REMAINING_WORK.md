# TVC — Remaining Work: Enterprise Production Readiness

**Document Version:** 1.0  
**Date:** February 20, 2026  
**Status:** Active Engineering Backlog  
**Audience:** Engineering Team, Technical Stakeholders

---

## Table of Contents

1. [Current State Summary](#1-current-state-summary)
2. [Sprint 4.1 — PII Detection & Redaction](#2-sprint-41--pii-detection--redaction)
3. [Sprint 4.2 — API Server: Real Handlers](#3-sprint-42--api-server-real-handlers)
4. [Sprint 4.3 — Dashboard Frontend](#4-sprint-43--dashboard-frontend)
5. [Sprint 4.4 — Authentication & Authorization](#5-sprint-44--authentication--authorization)
6. [Sprint 4.5 — Billing & Subscription Management](#6-sprint-45--billing--subscription-management)
7. [Redis Integration](#7-redis-integration)
8. [Database Hardening & Scaling](#8-database-hardening--scaling)
9. [Security Hardening](#9-security-hardening)
10. [Observability & Monitoring](#10-observability--monitoring)
11. [Testing — Closing the Gaps](#11-testing--closing-the-gaps)
12. [Performance Engineering](#12-performance-engineering)
13. [Infrastructure & Deployment](#13-infrastructure--deployment)
14. [Developer Experience & Distribution](#14-developer-experience--distribution)
15. [Compliance & Audit](#15-compliance--audit)
16. [Documentation](#16-documentation)
17. [Future Features](#17-future-features)
18. [Priority Matrix](#18-priority-matrix)
19. [Risk Register](#19-risk-register)

---

## 1. Current State Summary

### Completed (Sprints 1–3)

| Component | Status | Test Coverage | Notes |
|-----------|--------|---------------|-------|
| Go project structure | Done | — | Monorepo with `tvc-go/` and `tvc-frontend/` |
| JSON Diff Engine | Done | 18 tests | Deep recursive comparison, ignore paths, array modes |
| OpenAPI Schema Comparison | Done | — | Breaking change detection for 3.x specs |
| CLI (`tvc diff`, `tvc schema diff`, `tvc replay`) | Done | — | Cross-compiled, JSON + text output |
| Traffic Proxy | Done | 26 tests | Reverse proxy, middleware, capture pipeline, sampling |
| Replay Engine | Done | 32 tests | Worker pool, rate limiting, retries, comparison, reporting |
| PostgreSQL Storage | Done | — | Full CRUD, repository interface, migrations |
| Database Migrations | Done | — | Initial schema (orgs, projects, envs, traffic, replays, schemas) |
| Docker Compose | Done | — | Postgres 16 + Redis 7 |
| CI/CD Workflows | Done | — | GitHub Actions for Go (lint, test, build) and Frontend (build) |
| Frontend Landing Page | Done | — | Header, Hero, Stats, Features, HowItWorks, Capabilities, FAQ, CTA, Footer |

**Total backend tests: 76 | Test failures: 0 | Race conditions: 0**

### Not Started / Stubbed

| Component | Current State |
|-----------|---------------|
| API handlers | All 18 endpoints return `"not yet implemented"` |
| `cmd/replayer/main.go` | Placeholder — prints version string |
| PII detection | Not started |
| Redis usage | Config defined but not wired |
| Auth (backend) | No JWT validation, no middleware |
| Auth (frontend) | Supabase client exists, no auth flows or protected routes |
| Dashboard UI | No pages, no components, no API client |
| Billing (Stripe) | Not started |
| Dockerfiles | Not created |
| Integration/E2E tests | Not started |
| Frontend tests | Not started |

---

## 2. Sprint 4.1 — PII Detection & Redaction

**Priority:** Critical  
**Estimated Effort:** 4–5 days  
**Dependency:** None (can start immediately)

PII handling is the single biggest enterprise trust requirement. If captured traffic contains unmasked PII and that data is exposed, it's a regulatory and legal catastrophe.

### 2.1 Core Detection Engine

Create `internal/pii/` package:

| File | Purpose |
|------|---------|
| `detector.go` | Core detection engine — scans `interface{}` trees (JSON payloads) |
| `redactor.go` | Replaces matched PII with redaction tokens |
| `patterns.go` | Regex and heuristic pattern definitions |
| `config.go` | Configuration for enabling/disabling patterns, custom patterns |
| `detector_test.go` | Comprehensive tests with real-world payloads |
| `redactor_test.go` | Verify redaction output, edge cases |
| `benchmark_test.go` | Performance benchmarks |

### 2.2 Detection Patterns (Minimum Set)

| Pattern | Example | Regex Complexity | False Positive Risk |
|---------|---------|-----------------|-------------------|
| Email address | `user@example.com` | Low | Low |
| US phone number | `(555) 123-4567`, `555-123-4567` | Medium | Medium |
| International phone | `+44 20 7946 0958` | Medium | Medium |
| Credit card (Visa) | `4111 1111 1111 1111` | Low | Low |
| Credit card (Mastercard) | `5500 0000 0000 0004` | Low | Low |
| Credit card (Amex) | `3782 822463 10005` | Low | Low |
| SSN | `123-45-6789` | Low | Medium |
| US Driver License | State-specific formats | High | High |
| US Passport | 9-digit alphanumeric | Medium | Medium |
| Date of Birth | `1990-01-15`, `01/15/1990` | Medium | High |
| IPv4 address | `192.168.1.1` | Low | Low |
| IPv6 address | `2001:0db8:85a3::8a2e:0370:7334` | Medium | Low |
| JWT token | `eyJhbGci...` | Low | Very Low |
| API key / Bearer token | `sk_live_...`, `Bearer ...` | Medium | Low |
| AWS Access Key | `AKIA...` | Low | Very Low |
| Street address | `123 Main St, Apt 4B` | Very High | Very High |
| Name detection | Context-dependent | Very High | Very High |

### 2.3 Luhn Validation for Credit Cards

Raw regex for credit cards produces false positives. Implement the **Luhn algorithm** to validate any 13–19 digit sequence before flagging it as a credit card:

```
func luhnCheck(number string) bool
```

### 2.4 Redaction Strategy

- **Default mode:** Replace with pattern-specific tokens — `[EMAIL_REDACTED]`, `[PHONE_REDACTED]`, `[CC_REDACTED]`
- **Hash mode:** Replace with deterministic SHA-256 hash (preserves referential integrity across logs)
- **Mask mode:** Partial masking — `u***@example.com`, `****-****-****-1111`
- **Configurable per-project:** Some projects may want to keep IP addresses but redact emails
- Redaction must be **applied before storage** — the database should never see raw PII

### 2.5 Deep Payload Scanning

PII can appear anywhere:
- Request body (JSON objects, deeply nested)
- Response body
- Query parameters (`?email=user@example.com`)
- Headers (`Authorization: Bearer ...`, `X-Customer-Email: ...`)
- URL path segments (`/users/john.doe@email.com/orders`)

The scanner must recursively walk `map[string]interface{}` and `[]interface{}` trees, scanning every string value and every map key.

### 2.6 Integration with Proxy Capture Pipeline

The PII detector must be inserted into the capture pipeline **before** the database write:

```
Request → Proxy → [Capture] → [PII Scan + Redact] → Channel → Worker → Database
```

- Flag `pii_redacted = true` on the `TrafficLog` if any PII was found and redacted
- Log a structured summary of what was redacted (without logging the actual PII values)
- Track PII detection metrics (count per type, per endpoint)

### 2.7 Performance Requirements

| Metric | Target |
|--------|--------|
| Scan 1KB payload | < 1ms |
| Scan 10KB payload | < 5ms |
| Scan 100KB payload | < 50ms |
| Memory overhead per scan | < 2x payload size |

Use `sync.Pool` for regex match buffers to reduce GC pressure under high throughput.

### 2.8 Configuration

```yaml
pii:
  enabled: true
  mode: "redact"          # redact | hash | mask
  patterns:
    email: true
    phone: true
    credit_card: true
    ssn: true
    api_key: true
    jwt: true
    ip_address: false      # disabled by default
  custom_patterns:
    - name: "internal_id"
      regex: "CUST-[A-Z0-9]{10}"
      replacement: "[INTERNAL_ID_REDACTED]"
  scan_headers: true
  scan_query_params: true
  scan_url_path: true
  scan_request_body: true
  scan_response_body: true
```

### 2.9 Testing Requirements

- Unit tests for every pattern (true positives and true negatives)
- Boundary tests (numbers that look like credit cards but fail Luhn, dates that look like SSNs)
- Performance benchmarks with realistic payload sizes
- Integration test: full proxy → capture → redact → store pipeline
- Fuzz testing: random payloads to ensure no panics
- **Minimum 95% detection rate, < 5% false positive rate** for each pattern type

---

## 3. Sprint 4.2 — API Server: Real Handlers

**Priority:** Critical  
**Estimated Effort:** 6–8 days  
**Dependency:** PostgreSQL storage (done)

All 18 API endpoints in `internal/api/routes.go` currently return `"not yet implemented"`. This needs to be fully wired with validation, authentication, pagination, and proper error handling.

### 3.1 Architecture Refactor

The current `routes.go` uses plain handler functions. For enterprise quality, refactor to dependency-injected handler structs:

```
internal/api/
├── routes.go              # Route registration
├── middleware/
│   ├── auth.go            # JWT/Supabase auth middleware
│   ├── rate_limit.go      # Per-user/per-org rate limiting
│   ├── request_id.go      # Request ID propagation
│   ├── logging.go         # Structured request logging
│   ├── recovery.go        # Panic recovery
│   └── cors.go            # CORS configuration
├── handlers/
│   ├── projects.go        # Project CRUD
│   ├── traffic.go         # Traffic listing, stats
│   ├── replays.go         # Replay CRUD, start/stop
│   ├── schemas.go         # Schema upload, diff
│   ├── environments.go    # Environment management (NEW)
│   ├── organizations.go   # Organization management (NEW)
│   └── health.go          # Health + readiness probes
├── request/
│   ├── parser.go          # JSON body parsing with size limits
│   ├── validator.go       # Input validation
│   └── pagination.go      # Cursor-based pagination parser
├── response/
│   ├── json.go            # Standard JSON response writer
│   ├── error.go           # Error response formatting
│   └── pagination.go      # Pagination metadata in responses
└── dto/
    ├── project.go         # Request/response DTOs for projects
    ├── traffic.go         # Request/response DTOs for traffic
    ├── replay.go          # Request/response DTOs for replays
    └── schema.go          # Request/response DTOs for schemas
```

### 3.2 Endpoint Implementation Checklist

#### Projects

| Endpoint | Method | Auth | Validation | Pagination | Notes |
|----------|--------|------|------------|------------|-------|
| `/api/v1/projects` | GET | Org member | — | Cursor-based | Filter by org |
| `/api/v1/projects` | POST | Org admin | Name, slug, org_id | — | Slug uniqueness per org |
| `/api/v1/projects/{id}` | GET | Org member | UUID format | — | Include environment list |
| `/api/v1/projects/{id}` | PUT | Org admin | Name, description | — | Cannot change org |
| `/api/v1/projects/{id}` | DELETE | Org owner | UUID format | — | Cascade deletes traffic |

#### Traffic

| Endpoint | Method | Auth | Validation | Pagination | Notes |
|----------|--------|------|------------|------------|-------|
| `/api/v1/projects/{id}/traffic` | GET | Org member | Time range, method, path, status | Cursor-based | Index on (project_id, timestamp DESC) |
| `/api/v1/projects/{id}/traffic/{logId}` | GET | Org member | UUID format | — | Full request/response bodies |
| `/api/v1/projects/{id}/traffic/stats` | GET | Org member | Time range | — | Aggregate counts, avg latency, error rates |

#### Replays

| Endpoint | Method | Auth | Validation | Pagination | Notes |
|----------|--------|------|------------|------------|-------|
| `/api/v1/projects/{id}/replays` | GET | Org member | — | Cursor-based | List with status filter |
| `/api/v1/projects/{id}/replays` | POST | Org admin | Source/target env, filters, sample size | — | Validate environments exist |
| `/api/v1/projects/{id}/replays/{replayId}` | GET | Org member | UUID | — | Include summary stats |
| `/api/v1/projects/{id}/replays/{replayId}/start` | POST | Org admin | — | — | Async — returns 202 Accepted |
| `/api/v1/projects/{id}/replays/{replayId}/stop` | POST | Org admin | — | — | Graceful cancellation |
| `/api/v1/projects/{id}/replays/{replayId}/results` | GET | Org member | Severity filter | Cursor-based | Include diff reports |

#### Schemas

| Endpoint | Method | Auth | Validation | Pagination | Notes |
|----------|--------|------|------------|------------|-------|
| `/api/v1/projects/{id}/schemas` | GET | Org member | — | Cursor-based | Version history |
| `/api/v1/projects/{id}/schemas` | POST | Org admin | Schema content, version, type | — | Validate OpenAPI/GraphQL |
| `/api/v1/projects/{id}/schemas/diff` | POST | Org member | Two version IDs | — | Return diff report |

#### New Endpoints Needed

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/organizations` | GET/POST | Organization management |
| `/api/v1/organizations/{id}` | GET/PUT/DELETE | Single org management |
| `/api/v1/organizations/{id}/members` | GET/POST/DELETE | Member management |
| `/api/v1/projects/{id}/environments` | GET/POST | Environment management |
| `/api/v1/projects/{id}/environments/{envId}` | GET/PUT/DELETE | Single env management |
| `/api/v1/auth/me` | GET | Current user profile |
| `/api/v1/replays/{replayId}/export` | GET | PDF/CSV export |
| `/api/v1/health` | GET | Health check (no auth) |
| `/api/v1/ready` | GET | Readiness check (no auth) |

### 3.3 Standard Error Response Format

Every error response must follow this structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable description",
    "request_id": "req_abc123",
    "details": [
      { "field": "name", "message": "Name is required" },
      { "field": "sample_size", "message": "Must be between 1 and 10000" }
    ]
  }
}
```

Error codes: `VALIDATION_ERROR`, `NOT_FOUND`, `UNAUTHORIZED`, `FORBIDDEN`, `CONFLICT`, `RATE_LIMITED`, `INTERNAL_ERROR`, `SERVICE_UNAVAILABLE`

### 3.4 Pagination Strategy

Use cursor-based pagination (not offset-based) for scalable performance on large tables:

```
GET /api/v1/projects/{id}/traffic?limit=50&cursor=eyJ0cyI6MTcwOC4uLn0=

Response:
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJ0cyI6MTcwOC4uLn0=",
    "has_more": true,
    "total_estimate": 15420
  }
}
```

Cursors are base64-encoded JSON containing the sort key(s) from the last item.

### 3.5 Request Size Limits

| Resource | Max Body Size |
|----------|--------------|
| Project create/update | 10KB |
| Schema upload | 10MB |
| Replay create | 50KB |
| Default | 1MB |

Enforce via middleware — reject with `413 Payload Too Large` before any parsing.

### 3.6 Replayer Service Entry Point

`cmd/replayer/main.go` is currently a placeholder. Wire it up to:
1. Accept replay session IDs from the API (via database polling or Redis pub/sub)
2. Execute the replay using `internal/replayer/worker.go`
3. Update replay session status in the database
4. Send completion notifications

---

## 4. Sprint 4.3 — Dashboard Frontend

**Priority:** High  
**Estimated Effort:** 10–14 days  
**Dependencies:** API handlers (Sprint 4.2), Auth (Sprint 4.4)

The frontend currently has only a landing page. The entire authenticated dashboard needs to be built.

### 4.1 Missing Dependencies to Install

```
@tanstack/react-query          # Server state management
@tanstack/react-virtual        # Virtual scrolling for traffic tables
@radix-ui/react-dialog         # Modals
@radix-ui/react-dropdown-menu  # Dropdown menus
@radix-ui/react-tabs           # Tab navigation
@radix-ui/react-toast          # Toast notifications
@radix-ui/react-select         # Select dropdowns
@radix-ui/react-switch         # Toggle switches
@radix-ui/react-tooltip        # Tooltips
react-hook-form                # Form management
@hookform/resolvers            # Zod resolver for react-hook-form
zod                            # Schema validation
recharts                       # Charts and data visualization
date-fns                       # Date formatting
lucide-react                   # Icon system (already partially present)
sonner                         # Toast alternative (simpler API)
cmdk                           # Command palette
@tanstack/react-table          # Table component
```

### 4.2 Route Structure

```
app/
├── (auth)/                       # Protected route group
│   ├── layout.tsx                # Dashboard shell (sidebar + header)
│   ├── dashboard/
│   │   └── page.tsx              # Overview: stats, recent activity, charts
│   ├── traffic/
│   │   ├── page.tsx              # Traffic stream: filterable table
│   │   └── [id]/
│   │       └── page.tsx          # Single request detail view
│   ├── replay/
│   │   ├── page.tsx              # Replay sessions list
│   │   ├── new/
│   │   │   └── page.tsx          # Create replay form (multi-step)
│   │   └── [id]/
│   │       ├── page.tsx          # Replay results
│   │       └── report/
│   │           └── page.tsx      # Detailed diff report
│   ├── schemas/
│   │   ├── page.tsx              # Schema version history
│   │   └── diff/
│   │       └── page.tsx          # Schema diff viewer
│   ├── settings/
│   │   ├── page.tsx              # General settings
│   │   ├── team/
│   │   │   └── page.tsx          # Team/member management
│   │   ├── environments/
│   │   │   └── page.tsx          # Environment configuration
│   │   ├── billing/
│   │   │   └── page.tsx          # Billing + subscription
│   │   └── api-keys/
│   │       └── page.tsx          # API key management
│   └── audit/
│       └── page.tsx              # Audit log
├── (marketing)/                  # Public pages (existing landing)
│   └── page.tsx
├── login/
│   └── page.tsx                  # Login page
├── signup/
│   └── page.tsx                  # Signup page
└── api/
    └── webhooks/
        └── stripe/
            └── route.ts          # Stripe webhook handler
```

### 4.3 Component Inventory

#### Layout Components

| Component | Purpose | Complexity |
|-----------|---------|------------|
| `DashboardShell` | Sidebar + header + main content area | Medium |
| `Sidebar` | Navigation with collapsible sections, project switcher | Medium |
| `DashboardHeader` | Breadcrumbs, user menu, notifications | Medium |
| `ProjectSwitcher` | Dropdown to switch between projects | Low |
| `CommandPalette` | Cmd+K search across traffic, replays, schemas | High |

#### UI Components (Design System)

| Component | Source | Status |
|-----------|--------|--------|
| `Button` | Existing | Done |
| `Card` | Radix/custom | Needed |
| `Dialog` (Modal) | Radix | Needed |
| `DropdownMenu` | Radix | Needed |
| `Select` | Radix | Needed |
| `Tabs` | Radix | Needed |
| `Table` | TanStack Table | Needed |
| `Input` | Custom | Needed |
| `Textarea` | Custom | Needed |
| `Badge` | Custom | Needed |
| `Toast` | Radix/Sonner | Needed |
| `Tooltip` | Radix | Needed |
| `Switch` | Radix | Needed |
| `Skeleton` | Custom | Needed |
| `EmptyState` | Custom | Needed |
| `ErrorBoundary` | Custom | Needed |
| `LoadingSpinner` | Custom | Needed |
| `Pagination` | Custom (cursor-based) | Needed |
| `DateRangePicker` | Custom | Needed |
| `CodeViewer` | Custom (JSON syntax highlighting) | Needed |
| `CopyButton` | Custom | Needed |

#### Domain Components

| Component | Page | Complexity |
|-----------|------|------------|
| `StatsGrid` | Dashboard | Low |
| `TrafficChart` | Dashboard | Medium (Recharts) |
| `RecentActivity` | Dashboard | Low |
| `TrafficTable` | Traffic list | High (virtual scroll, filters, sort) |
| `TrafficFilterBar` | Traffic list | Medium (method, status, path, date range) |
| `RequestViewer` | Traffic detail | Medium (JSON viewer with syntax highlighting) |
| `ResponseViewer` | Traffic detail | Medium |
| `HeadersTable` | Traffic detail | Low |
| `MetadataPanel` | Traffic detail | Low |
| `ReplayConfigForm` | Replay new | High (multi-step form, environment selectors) |
| `ReplayProgress` | Replay detail | Medium (real-time progress bar) |
| `ReplayResultsSummary` | Replay detail | Medium (stats cards + severity breakdown) |
| `ReplayResultsTable` | Replay detail | High (filter by severity, expand diffs) |
| `DiffViewer` | Replay report | Very High (side-by-side JSON diff with highlighting) |
| `SeverityBadge` | Multiple | Low |
| `SchemaVersionList` | Schemas | Medium |
| `SchemaUploadForm` | Schemas | Medium |
| `SchemaDiffViewer` | Schema diff | High (breaking change highlights) |
| `TeamMemberList` | Settings | Medium |
| `InviteMemberForm` | Settings | Low |
| `EnvironmentCard` | Settings | Low |
| `BillingOverview` | Billing | Medium |
| `UsageChart` | Billing | Medium |
| `SubscriptionManager` | Billing | High (Stripe Elements) |
| `AuditLogTable` | Audit | Medium |

### 4.4 API Client Layer

Create `lib/api/`:

| File | Purpose |
|------|---------|
| `client.ts` | Base fetch wrapper with auth headers, error handling, retry logic |
| `projects.ts` | Project CRUD API calls |
| `traffic.ts` | Traffic listing and detail API calls |
| `replays.ts` | Replay CRUD, start/stop, results API calls |
| `schemas.ts` | Schema management API calls |
| `organizations.ts` | Organization and member management API calls |
| `environments.ts` | Environment management API calls |
| `billing.ts` | Subscription and usage API calls |

### 4.5 TanStack Query Hooks

Create `lib/hooks/`:

| Hook | Purpose |
|------|---------|
| `useProjects` | List projects with caching |
| `useProject` | Single project with auto-refetch |
| `useTraffic` | Paginated traffic with infinite scroll |
| `useTrafficLog` | Single traffic log detail |
| `useTrafficStats` | Dashboard stats (30s stale time, 60s refetch) |
| `useReplays` | Replay session list |
| `useReplay` | Single replay with polling when `status === 'running'` |
| `useReplayResults` | Paginated replay results |
| `useSchemas` | Schema version list |
| `useCreateReplay` | Mutation hook |
| `useStartReplay` | Mutation hook with optimistic update |
| `useUploadSchema` | Mutation hook with file upload |
| `useOrganization` | Current org context |
| `useMembers` | Team member list |

### 4.6 Zod Schemas for Validation

Create `lib/schemas/`:

| Schema | Fields to Validate |
|--------|--------------------|
| `createProjectSchema` | name (1–255), description (optional, max 1000), slug (alphanumeric + dashes) |
| `createReplaySchema` | name, sourceEnvironmentId (UUID), targetEnvironmentId (UUID), sampleSize (1–10000), trafficFilter (paths, methods, statusCodes, timeRange) |
| `uploadSchemaSchema` | version (semver), schemaType (openapi/graphql), content (valid JSON/YAML) |
| `inviteMemberSchema` | email (valid), role (admin/member) |
| `createEnvironmentSchema` | name (1–100), baseUrl (valid URL), isSource (boolean) |
| `trafficFilterSchema` | methods (array of HTTP methods), statusCodes (array 100–599), paths (array of strings), startDate, endDate |

### 4.7 State Management

- **Server state:** TanStack Query for all API data
- **Global UI state:** React Context for sidebar, theme, current project/org
- **Local state:** `useState` for component-level state (form inputs, expanded rows, filters)
- **URL state:** `nuqs` or `useSearchParams` for persisting filters in the URL (so users can share filtered views)

### 4.8 Performance Requirements

| Metric | Target |
|--------|--------|
| First Contentful Paint | < 1.5s |
| Time to Interactive | < 3s |
| Largest Contentful Paint | < 2.5s |
| Lighthouse Performance Score | > 90 |
| Traffic table render (1000 rows) | < 200ms (virtual scroll) |
| Dashboard initial load | < 2s |

Use `next/dynamic` for code splitting heavy components (DiffViewer, Recharts) and `@tanstack/react-virtual` for any list exceeding 100 items.

---

## 5. Sprint 4.4 — Authentication & Authorization

**Priority:** Critical  
**Estimated Effort:** 5–7 days  
**Dependencies:** API handlers (Sprint 4.2)

### 5.1 Backend Auth (Go)

#### JWT Validation Middleware

Create `internal/api/middleware/auth.go`:

- Extract `Authorization: Bearer <token>` header
- Validate JWT signature against Supabase JWT secret
- Extract `user_id`, `email`, `role` from claims
- Inject user context into `context.Context`
- Return `401 Unauthorized` for invalid/expired tokens
- Return `403 Forbidden` for insufficient permissions

#### Supabase Integration

- Verify JWTs against Supabase's JWKS endpoint (RS256)
- Support both access tokens and service role tokens
- Cache JWKS keys with TTL (refresh every 6 hours)
- Handle token refresh edge cases (token expired during long request)

#### Role-Based Access Control (RBAC)

| Role | Permissions |
|------|------------|
| Owner | Full access, can delete org, manage billing, transfer ownership |
| Admin | Create/delete projects, manage environments, start replays, invite members |
| Member | View projects, view traffic, view replay results (read-only) |
| Viewer | View-only access to dashboards (no API access to raw data) |

Implement as middleware that checks:
1. User belongs to the organization that owns the project
2. User's role in the organization meets the minimum required role
3. Project-level overrides (optional — future)

#### API Key Authentication

For CLI and CI/CD usage:
- Generate API keys scoped to a project or organization
- Store hashed keys in the database
- Support in both `Authorization: Bearer <api_key>` and `X-API-Key: <key>` headers
- Rate limit per API key
- Key rotation (create new, deprecate old with grace period)

### 5.2 Frontend Auth

#### Pages

| Page | Features |
|------|----------|
| `/login` | Email/password, magic link, OAuth (GitHub, Google) |
| `/signup` | Registration form with org creation |
| `/forgot-password` | Password reset flow |
| `/verify-email` | Email verification |
| `/invite/{token}` | Accept team invitation |

#### Auth Flow

1. User signs up → Creates Supabase Auth user → Creates organization → Redirects to dashboard
2. User logs in → Gets JWT → Stored in secure HttpOnly cookie (Supabase SSR handles this)
3. Every API call → Attach JWT from cookie → Go backend validates → Returns data
4. Token refresh → Supabase SSR middleware handles automatic refresh
5. Logout → Clear session → Redirect to landing

#### Protected Route Middleware

Create Next.js middleware (`middleware.ts`) that:
- Checks for valid Supabase session on all `/(auth)/` routes
- Redirects to `/login` if no session
- Redirects to `/dashboard` if authenticated user hits `/login`
- Handles token refresh transparently

### 5.3 Session Security

| Measure | Implementation |
|---------|---------------|
| Token storage | HttpOnly, Secure, SameSite=Lax cookies (NOT localStorage) |
| Token lifetime | Access: 1 hour, Refresh: 30 days |
| Concurrent sessions | Allow up to 5 active sessions per user |
| Session revocation | Revoke all sessions on password change |
| Brute force protection | Rate limit login attempts (5 per minute per IP) |
| Account lockout | Lock after 10 failed attempts, unlock after 30 minutes |

---

## 6. Sprint 4.5 — Billing & Subscription Management

**Priority:** High  
**Estimated Effort:** 5–7 days  
**Dependencies:** Auth (Sprint 4.4)

### 6.1 Stripe Integration (Backend)

#### Endpoints

| Endpoint | Purpose |
|----------|---------|
| `POST /api/v1/billing/checkout` | Create Stripe Checkout session |
| `POST /api/v1/billing/portal` | Create Stripe Customer Portal session |
| `GET /api/v1/billing/subscription` | Get current subscription details |
| `GET /api/v1/billing/usage` | Get current period usage |
| `POST /api/v1/webhooks/stripe` | Stripe webhook receiver |

#### Webhook Events to Handle

| Event | Action |
|-------|--------|
| `checkout.session.completed` | Create/activate subscription, update org tier |
| `customer.subscription.updated` | Update tier, limits, period dates |
| `customer.subscription.deleted` | Downgrade to free tier |
| `invoice.payment_failed` | Mark subscription as past_due, send warning |
| `invoice.payment_succeeded` | Clear past_due status |
| `customer.subscription.trial_will_end` | Send 3-day warning email |

#### Webhook Security

- Verify Stripe signature on every webhook (`stripe.ConstructEvent`)
- Idempotency: store processed event IDs, skip duplicates
- Retry handling: Stripe retries failed webhooks up to 3 days
- Dead letter queue: log unprocessable events for manual review

### 6.2 Subscription Tiers

| Feature | Free | Pro ($99/mo) | Enterprise ($499/mo) |
|---------|------|-------------|---------------------|
| CLI tool | Unlimited | Unlimited | Unlimited |
| Schema diffing | Unlimited | Unlimited | Unlimited |
| Traffic capture | — | 100K logs/month | Unlimited |
| Replay sessions | — | 50/month | Unlimited |
| Team members | 1 | 10 | Unlimited |
| Data retention | — | 30 days | 90 days |
| PII detection | — | Basic patterns | Custom patterns |
| API access | — | Standard rate | Priority rate |
| Support | Community | Email (48h SLA) | Dedicated (4h SLA) |
| Audit logs | — | — | Full audit trail |
| SSO/SAML | — | — | Included |
| Custom deployment | — | — | On-premise option |

### 6.3 Usage Tracking

Track and enforce:
- Traffic logs captured per billing period
- Replay sessions started per billing period
- Schema uploads per billing period
- API requests per hour (rate limiting)
- Storage consumed

#### Enforcement

- **Soft limit (80%):** Warning banner in dashboard, email notification
- **Hard limit (100%):** Block new traffic capture, block new replays
- **Grace period:** 24 hours after hitting hard limit before enforcement
- **Overage (Enterprise only):** Allow overage, bill at per-unit rate

### 6.4 Frontend Billing UI

| Component | Features |
|-----------|----------|
| `BillingOverview` | Current plan, usage bars, next billing date |
| `UsageBreakdown` | Per-resource usage with charts |
| `PlanSelector` | Compare plans, upgrade/downgrade |
| `PaymentMethod` | Stripe Elements card input |
| `InvoiceHistory` | Past invoices with PDF download links |
| `CancelSubscription` | Cancellation flow with feedback survey |

---

## 7. Redis Integration

**Priority:** High  
**Estimated Effort:** 3–4 days  
**Dependencies:** None (Redis is already in Docker Compose)

Redis is configured but not actually used anywhere. Wire it up for:

### 7.1 Traffic Capture Buffer

Currently the proxy writes directly to PostgreSQL via a channel buffer. Add Redis as an intermediate:

```
Capture → Redis List (LPUSH) → Worker Pool (BRPOP) → PostgreSQL
```

Benefits:
- Survives proxy restarts without data loss
- Decouples capture rate from database write speed
- Enables multiple proxy instances sharing one queue

### 7.2 Caching Layer

| Cache Key | TTL | Purpose |
|-----------|-----|---------|
| `project:{id}` | 5 min | Avoid repeated project lookups |
| `env:{id}` | 5 min | Avoid repeated environment lookups |
| `traffic:stats:{project}:{period}` | 30s | Dashboard stats (expensive aggregate queries) |
| `user:{id}:orgs` | 5 min | User organization memberships |
| `sub:{org}:limits` | 1 min | Subscription limits for enforcement |

### 7.3 Rate Limiting

Use Redis sorted sets for sliding window rate limiting:
- API rate limits per user/org/API key
- Login attempt limiting
- Replay start rate (prevent abuse)

### 7.4 Pub/Sub

- Replay completion notifications (API server → frontend via SSE/WebSocket)
- Configuration change propagation (hot-reload proxy config)
- Real-time traffic stream (proxy → dashboard via WebSocket)

### 7.5 Implementation

Create `internal/storage/redis.go`:

| Method | Purpose |
|--------|---------|
| `NewRedisStore(url string)` | Initialize connection pool |
| `EnqueueTraffic(log *TrafficLog)` | LPUSH to traffic queue |
| `DequeueTraffic(timeout)` | BRPOP from traffic queue |
| `SetCache(key, value, ttl)` | Generic cache set |
| `GetCache(key)` | Generic cache get |
| `IncrementRateLimit(key, window)` | Sliding window increment |
| `CheckRateLimit(key, limit, window)` | Check if under limit |
| `Publish(channel, message)` | Pub/sub publish |
| `Subscribe(channel)` | Pub/sub subscribe |

---

## 8. Database Hardening & Scaling

**Priority:** High  
**Estimated Effort:** 4–5 days

### 8.1 Table Partitioning (Traffic Logs)

The migration defines `traffic_logs` as `PARTITION BY RANGE (timestamp)` but **no partitions are actually created**. Without partitions, inserts will fail.

**Required:**
- Create a partition management function that auto-creates monthly partitions
- Run as a cron job or database function (triggered weekly, creates 2 months ahead)
- Implement partition pruning for old data (configurable retention)

```sql
CREATE OR REPLACE FUNCTION create_traffic_partition(start_date DATE)
RETURNS void AS $$
DECLARE
    partition_name TEXT;
    end_date DATE;
BEGIN
    partition_name := 'traffic_logs_' || to_char(start_date, 'YYYY_MM');
    end_date := start_date + INTERVAL '1 month';
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF traffic_logs
         FOR VALUES FROM (%L) TO (%L)',
        partition_name, start_date, end_date
    );
END;
$$ LANGUAGE plpgsql;
```

### 8.2 Index Optimization

**Missing indexes (add to migration 002):**

```sql
-- GIN index for JSONB full-text search on request/response bodies
CREATE INDEX idx_traffic_logs_request_body_gin ON traffic_logs USING gin(request_body);
CREATE INDEX idx_traffic_logs_response_body_gin ON traffic_logs USING gin(response_body);

-- Trigram index for LIKE/ILIKE queries on path
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX idx_traffic_logs_path_trgm ON traffic_logs USING gin(path gin_trgm_ops);

-- Composite index for common dashboard query patterns
CREATE INDEX idx_traffic_logs_project_method_status
    ON traffic_logs(project_id, method, status_code, timestamp DESC);

-- Partial index for error traffic (frequently queried)
CREATE INDEX idx_traffic_logs_errors
    ON traffic_logs(project_id, timestamp DESC)
    WHERE status_code >= 400;

-- Index for replay results by severity
CREATE INDEX idx_replay_results_severity_filter
    ON replay_results(replay_session_id, severity)
    WHERE severity IN ('error', 'breaking');
```

### 8.3 Connection Pooling

The current `PostgresStore` sets `MaxOpenConns(25)`. For production:

- Deploy **PgBouncer** as a connection pooler in front of PostgreSQL
- Set application-side pool to `MaxOpenConns(10)` (let PgBouncer manage the rest)
- Use `transaction` pooling mode in PgBouncer for web workloads
- Monitor `pg_stat_activity` for connection leaks

### 8.4 Data Retention & Archival

| Tier | Retention | Archival |
|------|-----------|----------|
| Free | N/A (no traffic capture) | — |
| Pro | 30 days | Automatic delete |
| Enterprise | 90 days live, 1 year archived | Move to S3 (Parquet format) |

**Implementation:**
- PostgreSQL function to drop partitions older than retention period
- Background job to export old partitions to S3 before dropping
- Separate "cold storage" query path for archived data

### 8.5 Query Performance Targets

| Query | Target | Current Status |
|-------|--------|----------------|
| Traffic list (paginated, filtered by project + time) | < 20ms | Untested |
| Traffic stats aggregate (1M rows) | < 100ms | Untested |
| Replay results list (paginated) | < 20ms | Untested |
| Schema version list | < 10ms | Untested |
| Full-text search on path | < 50ms | No index yet |

Run `EXPLAIN ANALYZE` on all queries during development and record baseline performance.

### 8.6 Migration Tooling

- Integrate `golang-migrate` into the Makefile and the binary itself
- `tvc-api migrate up` command
- `tvc-api migrate down` command
- `tvc-api migrate status` command
- Embed migration files in the Go binary using `embed.FS`

### 8.7 Backup & Recovery

- Automated daily backups (Supabase handles this, but verify retention)
- Point-in-time recovery capability
- Tested restore procedure (document and practice quarterly)
- Separate read replica for dashboard queries (prevents heavy reads from impacting writes)

---

## 9. Security Hardening

**Priority:** Critical  
**Estimated Effort:** 5–7 days (spread across multiple sprints)

### 9.1 Transport Security

| Measure | Details |
|---------|---------|
| TLS termination | All external traffic over HTTPS (TLS 1.2+, prefer 1.3) |
| HSTS header | `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload` |
| Certificate management | Auto-renewal via Let's Encrypt or AWS ACM |
| Internal mTLS | Mutual TLS between proxy → API → database (production) |

### 9.2 HTTP Security Headers

Add middleware to set on every response:

```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 0 (modern approach: rely on CSP)
Referrer-Policy: strict-origin-when-cross-origin
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' https://js.stripe.com; frame-src https://js.stripe.com; connect-src 'self' https://*.supabase.co;
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

### 9.3 Input Validation & Sanitization

| Layer | Validation |
|-------|-----------|
| Request body | JSON schema validation with size limits (reject > limit before parsing) |
| Path parameters | UUID format check, alphanumeric slug check |
| Query parameters | Whitelist allowed params, type check, range check |
| Headers | Strip unexpected headers before forwarding (proxy) |
| SQL | Parameterized queries only (already done — verify no raw string concatenation exists anywhere) |

### 9.4 Rate Limiting

| Endpoint Category | Limit | Window | Key |
|------------------|-------|--------|-----|
| Authentication | 5 requests | 1 minute | IP |
| API (Free tier) | 100 requests | 1 minute | API key |
| API (Pro tier) | 1000 requests | 1 minute | API key |
| API (Enterprise) | 10000 requests | 1 minute | API key |
| Webhook receivers | 100 requests | 1 minute | IP |
| Schema upload | 10 requests | 1 hour | User |
| Replay start | 5 requests | 1 minute | User |

Return `429 Too Many Requests` with `Retry-After` header.

### 9.5 Secrets Management

| Secret | Storage | Rotation |
|--------|---------|----------|
| Database credentials | Environment variables (never in config files) | 90 days |
| Redis credentials | Environment variables | 90 days |
| Supabase JWT secret | Environment variable | On compromise |
| Stripe API keys | Environment variable | On compromise |
| Stripe webhook secret | Environment variable | On rotation |
| API encryption keys | KMS (AWS/GCP) or Vault | Annual |

**Never log secrets.** Add a log scrubber that redacts known secret patterns from log output.

### 9.6 Data Encryption

| Data | At Rest | In Transit |
|------|---------|------------|
| Traffic logs (request/response bodies) | PostgreSQL TDE or column-level encryption | TLS |
| PII (redacted) | Encrypted JSONB columns (AES-256-GCM) | TLS |
| API keys | bcrypt hash (never store plaintext) | TLS |
| User passwords | Supabase handles (bcrypt) | TLS |
| Backup files | S3 SSE-KMS | TLS |

### 9.7 Audit Logging

Every security-relevant action must be logged:

| Event | Data Captured |
|-------|--------------|
| User login | user_id, ip, user_agent, success/failure, timestamp |
| User logout | user_id, timestamp |
| Failed login attempt | email, ip, user_agent, timestamp |
| API key created | user_id, key_id (not the key itself), scope, timestamp |
| API key revoked | user_id, key_id, timestamp |
| Project created/deleted | user_id, project_id, timestamp |
| Replay started | user_id, replay_id, project_id, timestamp |
| Member invited/removed | actor_id, target_email, org_id, role, timestamp |
| Subscription changed | org_id, old_tier, new_tier, timestamp |
| PII detected | project_id, traffic_log_id, pii_types, timestamp |
| Bulk data export | user_id, resource_type, count, timestamp |

Store in an append-only `audit_logs` table (never delete, never update).

### 9.8 Dependency Security

- Run `govulncheck` in CI to detect known Go vulnerabilities
- Run `npm audit` in CI for frontend vulnerabilities
- Enable Dependabot or Renovate for automated dependency updates
- Pin all dependency versions (already done in `go.mod` and `package-lock.json`)
- Review new dependencies before adding (check maintenance, license, CVE history)

### 9.9 CORS Configuration

The proxy middleware currently allows `*` origin. For production:

```go
AllowedOrigins:   []string{"https://app.tvc.dev", "https://tvc.dev"}
AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
AllowedHeaders:   []string{"Authorization", "Content-Type", "X-Request-ID"}
AllowCredentials: true
MaxAge:           86400
```

---

## 10. Observability & Monitoring

**Priority:** High  
**Estimated Effort:** 4–5 days

### 10.1 Structured Logging

The `zerolog` logger is in place. Ensure all logs include:

| Field | Purpose |
|-------|---------|
| `timestamp` | ISO 8601 with timezone |
| `level` | debug, info, warn, error, fatal |
| `service` | proxy, api, replayer, cli |
| `request_id` | Correlation ID across services |
| `user_id` | Authenticated user (if applicable) |
| `method` + `path` | HTTP request context |
| `status` | Response status code |
| `duration_ms` | Request processing time |
| `error` | Error message + stack trace (error level only) |

**Never log:** passwords, tokens, full request/response bodies (use PII-safe summaries), credit card numbers.

### 10.2 Metrics (Prometheus)

Expose `/metrics` endpoint (already partially in proxy). Add:

**Proxy metrics:**

| Metric | Type | Labels |
|--------|------|--------|
| `tvc_proxy_requests_total` | Counter | method, path, status, upstream |
| `tvc_proxy_request_duration_seconds` | Histogram | method, path |
| `tvc_proxy_capture_buffer_size` | Gauge | — |
| `tvc_proxy_capture_dropped_total` | Counter | reason (buffer_full, sampling) |
| `tvc_proxy_pii_detected_total` | Counter | type (email, phone, cc) |
| `tvc_proxy_upstream_errors_total` | Counter | upstream, error_type |

**API metrics:**

| Metric | Type | Labels |
|--------|------|--------|
| `tvc_api_requests_total` | Counter | method, path, status |
| `tvc_api_request_duration_seconds` | Histogram | method, path |
| `tvc_api_auth_failures_total` | Counter | reason (expired, invalid, missing) |

**Replayer metrics:**

| Metric | Type | Labels |
|--------|------|--------|
| `tvc_replayer_sessions_total` | Counter | status (completed, failed) |
| `tvc_replayer_requests_total` | Counter | result (match, mismatch, error) |
| `tvc_replayer_drift_score` | Histogram | — |
| `tvc_replayer_request_duration_seconds` | Histogram | — |

**Database metrics:**

| Metric | Type | Labels |
|--------|------|--------|
| `tvc_db_query_duration_seconds` | Histogram | operation (select, insert, update) |
| `tvc_db_connections_active` | Gauge | — |
| `tvc_db_connections_idle` | Gauge | — |

### 10.3 Distributed Tracing (OpenTelemetry)

Integrate OpenTelemetry SDK:
- Trace every request from ingress through proxy → API → database
- Propagate trace context via `traceparent` header
- Export to Jaeger, Zipkin, or Grafana Tempo
- Key spans: middleware execution, database queries, Redis operations, upstream proxy calls, replay HTTP calls

### 10.4 Alerting Rules

| Alert | Condition | Severity |
|-------|-----------|----------|
| Proxy p99 latency > 100ms | 5 consecutive minutes | Warning |
| Proxy p99 latency > 500ms | 2 consecutive minutes | Critical |
| Proxy error rate > 5% | 5 consecutive minutes | Critical |
| Capture buffer > 80% full | Immediate | Warning |
| Capture buffer > 95% full | Immediate | Critical |
| Database connection pool exhausted | Immediate | Critical |
| Replay session stuck in "running" > 1 hour | Immediate | Warning |
| API 5xx rate > 1% | 5 consecutive minutes | Critical |
| Disk usage > 80% | Immediate | Warning |
| Certificate expiring < 14 days | Daily check | Warning |

### 10.5 Health Checks

| Endpoint | Checks | Purpose |
|----------|--------|---------|
| `GET /health` | Process is alive | Load balancer liveness |
| `GET /ready` | DB connectable, Redis connectable, disk writable | Load balancer readiness |
| `GET /metrics` | — | Prometheus scraping |

Return structured health response:

```json
{
  "status": "healthy",
  "checks": {
    "database": { "status": "up", "latency_ms": 2 },
    "redis": { "status": "up", "latency_ms": 1 },
    "disk": { "status": "up", "free_gb": 45.2 }
  },
  "version": "1.2.0",
  "uptime_seconds": 86420
}
```

---

## 11. Testing — Closing the Gaps

**Priority:** Critical  
**Estimated Effort:** 8–10 days (ongoing)

### 11.1 Current Test Coverage

| Package | Tests | Coverage | Target |
|---------|-------|----------|--------|
| `internal/diffing` | 18 | ~85% | 90% |
| `internal/proxy` | 26 | ~75% | 85% |
| `internal/replayer` | 32 | ~80% | 85% |
| `internal/storage` | 0 | 0% | 80% |
| `internal/api` | 0 | 0% | 80% |
| `internal/config` | 0 | 0% | 70% |
| `internal/cli` | 0 | 0% | 70% |
| `internal/pii` | 0 (doesn't exist) | — | 95% |
| `pkg/errors` | 0 | 0% | 90% |
| `pkg/logger` | 0 | 0% | 70% |
| **Frontend** | 0 | 0% | 70% |

### 11.2 Integration Tests Needed

| Test Suite | What It Covers |
|-----------|----------------|
| `test/integration/proxy_capture_test.go` | Full proxy → capture → store pipeline with real Postgres |
| `test/integration/replay_flow_test.go` | Create replay → execute → compare → store results with real services |
| `test/integration/api_crud_test.go` | Full CRUD cycle through API handlers with real database |
| `test/integration/auth_flow_test.go` | Login → get token → access protected endpoint → unauthorized access |
| `test/integration/pii_pipeline_test.go` | Proxy captures traffic with PII → redaction → verify database has no PII |
| `test/integration/schema_diff_test.go` | Upload two schemas → compare → verify diff results |
| `test/integration/billing_webhook_test.go` | Simulate Stripe webhooks → verify subscription state changes |

Use `testcontainers-go` to spin up PostgreSQL and Redis containers per test suite.

### 11.3 End-to-End Tests

| Test | Flow |
|------|------|
| User onboarding | Sign up → Create org → Create project → Add environment → See empty dashboard |
| Traffic capture | Configure proxy → Send requests → Verify traffic appears in dashboard |
| Replay full flow | Select traffic → Configure replay → Start → Monitor progress → View results |
| Schema diff | Upload v1 schema → Upload v2 schema → Compare → See breaking changes |
| Billing upgrade | Free → Click upgrade → Stripe Checkout → Return → Verify Pro features |

Use Playwright for browser-based E2E tests.

### 11.4 Load & Stress Tests

| Test | Tool | Scenario | Target |
|------|------|----------|--------|
| Proxy throughput | `k6` or `hey` | 10,000 RPS for 5 minutes | < 5ms p95 added latency |
| Proxy under memory pressure | `k6` | 1,000 RPS with 100KB payloads | No OOM, graceful degradation |
| Replay throughput | Custom harness | 1,000 concurrent replays | Complete within 60s |
| API under load | `k6` | 500 RPS mixed read/write | < 100ms p95 |
| Database under load | `pgbench` | 10,000 concurrent connections | No connection exhaustion |
| Traffic table with 10M rows | SQL benchmark | Paginated queries | < 50ms |

### 11.5 Security Tests

| Test | Tool | Coverage |
|------|------|----------|
| SQL injection | `sqlmap` or manual | All API endpoints accepting user input |
| XSS | Manual + OWASP ZAP | All frontend forms and rendered content |
| CSRF | Manual | All state-changing API endpoints |
| Auth bypass | Manual | All protected endpoints, token manipulation |
| Rate limit bypass | Manual | Header spoofing, distributed requests |
| PII leakage | Custom scanner | Search database for unredacted PII patterns |
| Dependency vulnerabilities | `govulncheck`, `npm audit` | All dependencies |
| Secret scanning | `gitleaks` | Full git history |

### 11.6 Chaos Testing (Enterprise Readiness)

| Scenario | Expected Behavior |
|----------|-------------------|
| Kill PostgreSQL mid-request | Proxy continues forwarding, capture buffers in Redis, API returns 503 |
| Kill Redis mid-request | Proxy continues forwarding, capture falls back to direct DB write |
| Network partition (proxy ↔ upstream) | Return 502 to client, log error, don't crash |
| Disk full | Logs rotate, capture pauses, health check reports unhealthy |
| Memory pressure (90%+) | GC handles it, no OOM kill (monitor with cgroups) |
| Clock skew | JWT validation handles ±5s skew gracefully |

### 11.7 Frontend Testing

| Layer | Tool | Coverage |
|-------|------|----------|
| Component unit tests | Vitest + React Testing Library | All UI components |
| Hook tests | Vitest | All custom hooks |
| Integration tests | Vitest + MSW (Mock Service Worker) | Full page renders with mocked API |
| Visual regression | Chromatic or Percy | Key pages (dashboard, traffic, replay) |
| Accessibility | axe-core + Lighthouse | All pages (WCAG 2.1 AA) |
| E2E | Playwright | Critical flows (auth, replay, billing) |

---

## 12. Performance Engineering

**Priority:** High  
**Estimated Effort:** 5–7 days (iterative)

### 12.1 Go Performance Optimizations

| Area | Optimization | Impact |
|------|-------------|--------|
| **Memory allocation** | Use `sync.Pool` for frequently allocated objects (diff results, HTTP request buffers) | Reduce GC pauses |
| **JSON parsing** | Use `jsoniter` or `sonic` instead of `encoding/json` for hot paths | 2–5x faster parsing |
| **String building** | Use `strings.Builder` instead of `fmt.Sprintf` in loops | Reduce allocations |
| **HTTP client** | Connection pooling, keep-alive, disable redirect follow | Reduce replay latency |
| **Buffer sizing** | Profile and tune channel buffer sizes based on actual throughput | Prevent backpressure |
| **Context propagation** | Ensure all long-running operations respect `context.Context` cancellation | Clean shutdown |
| **Goroutine leaks** | Use `goleak` in tests to detect leaked goroutines | Prevent memory leaks |

### 12.2 Database Performance

| Optimization | Details |
|-------------|---------|
| Prepared statements | Use `db.Prepare` for frequently executed queries |
| Batch inserts | Batch traffic log inserts (100 at a time instead of 1-by-1) |
| Query result streaming | Use `sql.Rows` with streaming instead of loading full result sets |
| VACUUM scheduling | Configure autovacuum for traffic_logs (high-churn table) |
| Statistics target | Increase `default_statistics_target` for traffic_logs columns |
| pg_stat_statements | Enable and monitor slow queries |

### 12.3 Frontend Performance

| Optimization | Implementation |
|-------------|----------------|
| Code splitting | `next/dynamic` for DiffViewer, Charts, Monaco editor |
| Image optimization | `next/image` with WebP/AVIF |
| Bundle analysis | `@next/bundle-analyzer` — identify and eliminate large imports |
| Virtual scrolling | `@tanstack/react-virtual` for traffic table (1000+ rows) |
| Debounced search | 300ms debounce on filter inputs |
| Stale-while-revalidate | TanStack Query with `staleTime: 30000` for dashboard stats |
| Prefetching | `router.prefetch` for likely navigation targets |
| Service Worker | Cache static assets, offline support for docs |
| Web Workers | Offload JSON diff computation to a web worker (large payloads) |

### 12.4 Caching Strategy

| Layer | Tool | TTL | Invalidation |
|-------|------|-----|-------------|
| CDN | Vercel/CloudFront | Static: 1 year, API: no-cache | Deploy-based |
| API response | Redis | 30s–5min | Write-through on mutation |
| Database query | In-memory (Go) | 10s | LRU with max size |
| Frontend | TanStack Query | 30s stale | Refetch on focus, manual invalidation |

### 12.5 Benchmarking Suite

Create `test/benchmarks/`:

| Benchmark | What It Measures |
|-----------|-----------------|
| `BenchmarkDiffEngine_1KB` | Diff 1KB JSON objects |
| `BenchmarkDiffEngine_1MB` | Diff 1MB JSON objects |
| `BenchmarkDiffEngine_10MB` | Diff 10MB JSON objects |
| `BenchmarkPIIDetector_1KB` | PII scan 1KB payload |
| `BenchmarkPIIDetector_100KB` | PII scan 100KB payload |
| `BenchmarkProxyLatency` | End-to-end proxy overhead |
| `BenchmarkReplayThroughput` | Requests per second |
| `BenchmarkJSONMarshal` | Serialization performance |
| `BenchmarkDatabaseInsert` | Single insert latency |
| `BenchmarkDatabaseBatchInsert` | Batch insert throughput |

Run benchmarks in CI and track regressions with `benchstat`.

---

## 13. Infrastructure & Deployment

**Priority:** High  
**Estimated Effort:** 5–7 days

### 13.1 Dockerfiles

Create optimized multi-stage Dockerfiles:

| Dockerfile | Base | Final Size Target |
|-----------|------|-------------------|
| `Dockerfile.cli` | `golang:1.22-alpine` → `alpine:3.19` | < 20MB |
| `Dockerfile.proxy` | `golang:1.22-alpine` → `alpine:3.19` | < 25MB |
| `Dockerfile.api` | `golang:1.22-alpine` → `alpine:3.19` | < 25MB |
| `Dockerfile.replayer` | `golang:1.22-alpine` → `alpine:3.19` | < 25MB |
| `Dockerfile.frontend` | `node:20-alpine` → `node:20-alpine` (slim) | < 200MB |

Each Dockerfile must:
- Use non-root user
- Set `HEALTHCHECK` instruction
- Use `.dockerignore` to minimize build context
- Pin base image SHA256 digests (not just tags)
- Include security scanning (Trivy) in CI

### 13.2 Docker Compose (Enhanced)

Extend the existing `docker-compose.yml` for full local development:

```yaml
services:
  postgres: (existing)
  redis: (existing)
  proxy:      # tvc-go proxy service
  api:        # tvc-go API service
  replayer:   # tvc-go replayer service
  frontend:   # tvc-frontend dev server
  pgbouncer:  # connection pooler
  mailhog:    # email testing (for auth flows)
```

Add `docker-compose.test.yml` for integration test dependencies.

### 13.3 Kubernetes Manifests

Create `deployments/kubernetes/`:

| Resource | Purpose |
|----------|---------|
| `namespace.yaml` | Dedicated namespace |
| `proxy-deployment.yaml` | Proxy deployment (HPA: 2–10 replicas) |
| `api-deployment.yaml` | API deployment (HPA: 2–5 replicas) |
| `replayer-deployment.yaml` | Replayer deployment (HPA: 1–3 replicas) |
| `frontend-deployment.yaml` | Frontend deployment (2 replicas) |
| `proxy-service.yaml` | ClusterIP + LoadBalancer |
| `api-service.yaml` | ClusterIP |
| `ingress.yaml` | Ingress with TLS termination |
| `configmap.yaml` | Shared configuration |
| `secrets.yaml` | Encrypted secrets |
| `hpa.yaml` | Horizontal Pod Autoscaler rules |
| `pdb.yaml` | Pod Disruption Budgets |
| `network-policy.yaml` | Network segmentation |
| `resource-quotas.yaml` | CPU/memory limits |

### 13.4 Terraform / Infrastructure as Code

Create `deployments/terraform/`:

| Module | Resources |
|--------|-----------|
| `networking` | VPC, subnets, security groups, NAT gateway |
| `database` | Supabase project OR RDS PostgreSQL with read replicas |
| `cache` | ElastiCache Redis cluster |
| `compute` | EKS cluster OR ECS services |
| `cdn` | CloudFront distribution for frontend |
| `dns` | Route53 records |
| `monitoring` | CloudWatch alarms, Grafana dashboards |
| `storage` | S3 buckets for traffic archival, backups |

### 13.5 CI/CD Pipeline Enhancements

| Stage | Current | Needed |
|-------|---------|--------|
| Lint | Go + Frontend | Add `govulncheck`, `npm audit`, Trivy |
| Test | Unit only | Add integration, E2E, load tests |
| Build | Binary + frontend | Add Docker image build + push |
| Security | None | SAST (CodeQL), dependency scanning, secret scanning |
| Deploy (staging) | None | Auto-deploy `develop` branch to staging |
| Deploy (production) | None | Manual approval → deploy `main` to production |
| Rollback | None | One-click rollback to previous version |
| Notifications | None | Slack/Discord on deploy success/failure |
| Changelog | None | Auto-generate from conventional commits |

### 13.6 Environment Configuration

| Environment | Purpose | Infra |
|------------|---------|-------|
| Local | Developer machine | Docker Compose |
| CI | Automated testing | GitHub Actions |
| Staging | Pre-production validation | Kubernetes (small) |
| Production | Live service | Kubernetes (auto-scaling) |

Each environment needs its own:
- Database instance
- Redis instance
- Supabase project (or shared with row-level security)
- Stripe account (test mode for non-production)
- Domain and TLS certificate

---

## 14. Developer Experience & Distribution

**Priority:** Medium  
**Estimated Effort:** 3–5 days

### 14.1 CLI Distribution

| Channel | Implementation |
|---------|---------------|
| **Homebrew** | Create `homebrew-tvc` tap repository with formula |
| **npm** | Publish wrapper package (`npx tvc diff ...`) |
| **GitHub Releases** | Auto-publish binaries on tag push (already partially in CI) |
| **Docker** | `docker run ghcr.io/tvc-org/tvc:latest diff ...` |
| **Go install** | `go install github.com/tvc-org/tvc/cmd/cli@latest` |
| **curl installer** | `curl -sSf https://install.tvc.dev | sh` |

### 14.2 GitHub Action

Create `tvc-org/tvc-action` for marketplace:

```yaml
- uses: tvc-org/tvc-action@v1
  with:
    command: schema diff
    old-schema: openapi-v1.yaml
    new-schema: openapi-v2.yaml
    fail-on-breaking: true
    api-key: ${{ secrets.TVC_API_KEY }}
```

### 14.3 SDK / Client Libraries

| Language | Purpose |
|----------|---------|
| Go client | `tvc-go-sdk` — programmatic access to TVC API |
| TypeScript client | `@tvc-org/sdk` — type-safe API client for Node.js |
| Python client | `tvc-python` — for data science / analytics use cases |

Auto-generate from OpenAPI spec of our own API.

### 14.4 Documentation Site

Build a docs site (Nextra, Docusaurus, or Mintlify):

| Section | Content |
|---------|---------|
| Getting Started | Install CLI, first diff, connect to dashboard |
| CLI Reference | All commands, flags, exit codes, examples |
| Proxy Setup | Deployment modes, configuration, performance tuning |
| Replay Guide | Creating sessions, interpreting results, CI integration |
| API Reference | OpenAPI spec, authentication, rate limits |
| Self-Hosting | Docker Compose, Kubernetes, Terraform |
| Security | PII handling, encryption, compliance |
| Troubleshooting | Common issues and solutions |

---

## 15. Compliance & Audit

**Priority:** High (Enterprise)  
**Estimated Effort:** 5–7 days

### 15.1 Audit Trail

Create `audit_logs` table:

```sql
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    user_id UUID,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Append-only: no UPDATE or DELETE triggers
CREATE INDEX idx_audit_logs_org_time ON audit_logs(organization_id, timestamp DESC);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id, timestamp DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action, timestamp DESC);
```

### 15.2 Report Generation

| Report | Format | Trigger |
|--------|--------|---------|
| Replay compliance report | PDF | On-demand + after each replay |
| Schema change report | PDF | After each schema comparison |
| Traffic audit report | CSV | On-demand (date range, project filter) |
| PII detection report | PDF | Weekly automated + on-demand |
| Access audit report | CSV | On-demand (who accessed what, when) |

Use a Go PDF library (`jung-kurt/gofpdf` or `signintech/gopdf`) for server-side generation.

### 15.3 Data Export

| Export | Format | Features |
|--------|--------|----------|
| Traffic logs | JSON, CSV | Filtered by time, method, path, status |
| Replay results | JSON, CSV, PDF | Filtered by session, severity |
| Schema diffs | JSON, PDF | Full diff report with breaking changes |
| Audit logs | CSV | Filtered by date, user, action type |

All exports must:
- Respect PII redaction (never export raw PII)
- Include metadata (who requested, when, filter criteria)
- Be logged in the audit trail
- Be rate limited (1 export per minute per user)

### 15.4 Compliance Standards

| Standard | Relevance | Requirements |
|----------|-----------|-------------|
| SOC 2 Type II | Enterprise customers | Audit trail, access controls, encryption, incident response |
| GDPR | EU customers | Data minimization, right to deletion, data processing agreements |
| HIPAA | Healthcare customers | Encryption at rest + in transit, access controls, audit logging |
| PCI DSS | Payment data | Never store raw card numbers, encryption, access logging |

For GDPR specifically:
- Implement "right to be forgotten" — delete all traffic logs for a user/organization
- Data processing agreement template
- Cookie consent for the dashboard
- Data residency option (EU-only data storage)

---

## 16. Documentation

**Priority:** Medium  
**Estimated Effort:** 3–5 days

### 16.1 Missing Documentation

| Document | Status | Priority |
|----------|--------|----------|
| API Reference (OpenAPI spec for TVC's own API) | Not started | High |
| CLI User Guide | Not started | High |
| Proxy Deployment Guide | Not started | High |
| Self-Hosting Guide | Not started | Medium |
| Architecture Decision Records (ADRs) | Not started | Medium |
| Incident Response Playbook | Not started | High (Enterprise) |
| Security Whitepaper | Not started | High (Enterprise) |
| Contribution Guide | Not started | Low |
| Changelog | Not started | Medium |

### 16.2 Code Documentation

| Area | Status |
|------|--------|
| Go godoc comments on exported functions | Partial — some packages documented, others not |
| TypeScript JSDoc on component props | Not started |
| Package-level README files | Not started |
| Architecture diagram (C4 model) | Not started |
| Data flow diagrams | Not started |
| Sequence diagrams (key flows) | Not started |

### 16.3 Runbooks

Create operational runbooks for:

| Scenario | Content |
|----------|---------|
| Database migration failed | Steps to rollback, verify data, retry |
| Proxy not capturing traffic | Check config, verify upstream, check buffer, check DB |
| Replay session stuck | How to diagnose, force-cancel, re-run |
| High memory usage | Profile with pprof, check goroutine count, check buffer sizes |
| Certificate expiry | Renewal steps, verification, rollback |
| Database full | Identify largest tables, run partition pruning, expand storage |
| Redis connection failure | Failover steps, verify sentinels, check network |

---

## 17. Future Features

**Priority:** Backlog  
**Estimated Effort:** Varies

### 17.1 GraphQL Support

The product docs mention GraphQL support. Current implementation is REST/OpenAPI only.

| Task | Effort |
|------|--------|
| GraphQL schema parser (`vektah/gqlparser`) | 3 days |
| GraphQL breaking change rules | 2 days |
| GraphQL traffic capture (parse query from request body) | 2 days |
| GraphQL-specific diff viewer in dashboard | 3 days |

### 17.2 gRPC Support

| Task | Effort |
|------|--------|
| Protobuf schema comparison | 3 days |
| gRPC proxy (HTTP/2 aware) | 5 days |
| gRPC traffic capture and replay | 5 days |

### 17.3 WebSocket Support

| Task | Effort |
|------|--------|
| WebSocket message capture | 3 days |
| WebSocket replay (stateful) | 5 days |
| WebSocket diff viewer | 3 days |

### 17.4 Advanced Diffing

| Feature | Description |
|---------|-------------|
| Semantic diffing | Understand that `10` and `10.00` are semantically equivalent |
| Custom diff rules | User-defined ignore patterns, tolerance thresholds |
| Machine learning | Train a model to classify "expected drift" vs "real bugs" |
| Array matching | Smart array element matching (by ID field, not by index) |

### 17.5 Real-Time Dashboard

| Feature | Implementation |
|---------|---------------|
| WebSocket traffic stream | Server-sent events or WebSocket from API → frontend |
| Live replay progress | WebSocket with per-request updates |
| Alerting in dashboard | Toast notifications for breaking changes, high drift |

### 17.6 Multi-Tenancy

| Feature | Description |
|---------|-------------|
| Workspace isolation | Separate databases per enterprise customer |
| Custom domains | `api.customer.com` instead of `api.tvc.dev` |
| White-labeling | Customer branding on dashboard |
| On-premise deployment | Helm chart + operator for customer's Kubernetes |

### 17.7 ClickHouse Migration

When traffic_logs exceeds 100M rows, migrate high-volume queries to ClickHouse:

| Phase | Work |
|-------|------|
| Phase 1 | Dual-write to Postgres + ClickHouse |
| Phase 2 | Read from ClickHouse for analytics, Postgres for CRUD |
| Phase 3 | Drop Postgres traffic_logs, keep only ClickHouse |

---

## 18. Priority Matrix

### Immediate (This Sprint — Week 7)

| # | Task | Effort | Blocker |
|---|------|--------|---------|
| 1 | PII Detection Engine | 4 days | None |
| 2 | Wire up API handlers (projects, traffic, environments) | 4 days | None |
| 3 | Auth middleware (Go backend) | 2 days | None |

### Next Sprint (Week 8)

| # | Task | Effort | Blocker |
|---|------|--------|---------|
| 4 | Wire up remaining API handlers (replays, schemas) | 3 days | #2 |
| 5 | Frontend auth flow (login, signup, protected routes) | 3 days | #3 |
| 6 | Dashboard layout (sidebar, header, navigation) | 2 days | #5 |
| 7 | Redis integration (capture buffer, caching) | 3 days | None |

### Short-Term (Weeks 9–10)

| # | Task | Effort | Blocker |
|---|------|--------|---------|
| 8 | Dashboard — Traffic stream page | 3 days | #2, #6 |
| 9 | Dashboard — Traffic detail page | 2 days | #8 |
| 10 | Dashboard — Replay interface | 4 days | #4, #6 |
| 11 | Dashboard — Schema management | 2 days | #4, #6 |
| 12 | Dockerfiles (all services) | 2 days | None |
| 13 | Integration tests | 4 days | #2, #3 |
| 14 | Database partitioning (auto-create) | 1 day | None |

### Medium-Term (Weeks 11–12)

| # | Task | Effort | Blocker |
|---|------|--------|---------|
| 15 | Stripe billing integration | 5 days | #5 |
| 16 | Observability (metrics, tracing) | 4 days | None |
| 17 | Security hardening (rate limiting, headers, CORS) | 3 days | #3 |
| 18 | Frontend tests (unit + integration) | 4 days | #8, #10 |
| 19 | Load testing suite | 3 days | #2, #7 |
| 20 | Replayer service entry point (`cmd/replayer`) | 2 days | #4 |

### Long-Term (Weeks 13+)

| # | Task | Effort | Blocker |
|---|------|--------|---------|
| 21 | Audit trail + compliance reports | 4 days | #3 |
| 22 | Data export (CSV, PDF) | 3 days | #2 |
| 23 | CLI distribution (Homebrew, npm, Docker) | 2 days | None |
| 24 | GitHub Action | 2 days | #23 |
| 25 | Documentation site | 5 days | — |
| 26 | Kubernetes manifests | 3 days | #12 |
| 27 | Terraform modules | 5 days | #26 |
| 28 | GraphQL support | 10 days | — |
| 29 | E2E test suite (Playwright) | 5 days | #10, #15 |
| 30 | SOC 2 / GDPR readiness | 10 days | #21 |

---

## 19. Risk Register

| # | Risk | Probability | Impact | Mitigation |
|---|------|-------------|--------|------------|
| R1 | PII leakage in stored traffic | Low | Critical | Redact before storage, encrypt at rest, audit access |
| R2 | Proxy adds > 10ms latency | Medium | High | Profile hot paths, reduce allocations, async-only capture |
| R3 | Database growth overwhelms Postgres | High | High | Partitioning, retention policies, ClickHouse migration path |
| R4 | Replay overwhelms target server | Medium | Medium | Rate limiting, circuit breaker, dry-run mode |
| R5 | JWT validation bypass | Low | Critical | Use vetted libraries, test edge cases, rotate secrets |
| R6 | Stripe webhook replay attack | Low | High | Verify signatures, idempotency keys, event deduplication |
| R7 | Goroutine leak under error conditions | Medium | Medium | Use `goleak` in tests, context cancellation, timeouts everywhere |
| R8 | CORS misconfiguration exposes API | Low | High | Strict allowlist, test in production config |
| R9 | Dependency CVE | Medium | Medium | Automated scanning, Dependabot, rapid patching SLA (48h) |
| R10 | Migration failure corrupts data | Low | Critical | Always test migrations on a copy first, blue-green deployments |
| R11 | Redis failure loses buffered traffic | Medium | Medium | Persistent Redis (AOF), fallback to direct DB write |
| R12 | Certificate expiry causes outage | Low | High | Auto-renewal, 14-day alerting, manual renewal runbook |
| R13 | Secret committed to git | Low | Critical | Pre-commit hooks (`gitleaks`), secret scanning in CI |
| R14 | DDoS on proxy endpoint | Medium | High | Cloud WAF, connection limits, geographic filtering |
| R15 | Frontend XSS via traffic viewer | Medium | High | Sanitize all rendered content, CSP headers, no `dangerouslySetInnerHTML` |

---

## Summary

**Total remaining engineering effort: ~80–100 person-days** (conservative estimate)

The path to enterprise readiness requires disciplined execution across four domains:

1. **Feature completion** — PII, API handlers, dashboard, auth, billing (~40 days)
2. **Quality & testing** — Integration tests, E2E, load tests, security tests (~15 days)
3. **Infrastructure** — Docker, Kubernetes, CI/CD, observability (~15 days)
4. **Hardening** — Security, compliance, performance tuning, documentation (~20 days)

The foundation (Sprints 1–3) is solid. The diffing engine, proxy, and replay engine are well-tested and performant. The remaining work is primarily about wrapping these capabilities in production-grade packaging: authentication, authorization, billing, monitoring, and a polished user interface.

---

**Document Maintainer:** Engineering Team  
**Last Updated:** February 20, 2026  
**Next Review:** March 6, 2026
