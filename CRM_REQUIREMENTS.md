# CRM Service Requirements (for Platform Console Integration)

**Target repo:** https://github.com/SalehAlobaylan/CRM-Service

**Goal:** Refactor and scale the CRM service into a production-ready, API-first CRM backend that can be cleanly integrated into the existing **Platform Console** (admin dashboard) as a module, using HTTPS APIs and admin authentication.

---

## Implementation Status

| Module | Status | Notes |
|--------|--------|-------|
| Gin Framework | ✅ Complete | Replaced Gorilla Mux with Gin |
| PostgreSQL + GORM | ✅ Complete | Full persistence layer |
| JWT Auth (HS256 Verifier) | ✅ Complete | Middleware validates CMS-issued tokens |
| RBAC (admin/manager/agent) | ✅ Complete | Role-based access control |
| CORS Middleware | ✅ Complete | Configured for Vercel origins |
| Customers CRUD | ✅ Complete | With pagination, filtering, soft delete |
| Contacts CRUD | ✅ Complete | Nested under customers, primary designation |
| Deals CRUD | ✅ Complete | Pipeline stages, transitions |
| Activities CRUD | ✅ Complete | Including `/admin/me/activities` |
| Tags CRUD | ✅ Complete | With customer assignment |
| Reports Overview | ✅ Complete | `/admin/reports/overview` |
| Health/Ready/Metrics | ✅ Complete | Prometheus metrics |
| Structured Logging | ✅ Complete | Zap JSON logs with request IDs |
| Docker + Compose | ✅ Complete | Multi-stage build, PostgreSQL |
| SQL Migrations | ✅ Complete | golang-migrate compatible |
| **Notes CRUD** | ⚠️ Partial | Model exists, **endpoints not implemented** |
| **Audit Read Endpoint** | ⚠️ Partial | Logs written, **GET endpoint not implemented** |
| Attachments/File Upload | ❌ Not Started | Optional for v1 |

---

## Platform Console Context (must match)

### Console identity
- **Name:** Platform Console
- **Type:** Internal / Admin-only Web App (Control surface for platform + CRM operations).
- **Deployment:** Standalone Next.js app (Vercel) calling backend services over HTTPS.

### Services Console talks to
- **CMS (Platform):** Admin actions + observability via CMS APIs.
- **CRM Service:** CRM workflows and data via CRM APIs.

### Auth decision (MVP)
- **One shared login (SSO-like) with one JWT issuer**: Console authenticates once via **CMS** and uses the same token to call both CMS and CRM.
- **JWT transport:** `Authorization: Bearer <token>` header on every request.
- **JWT signing/verification:** **HS256** with a shared `JWT_SECRET` configured identically in CMS and CRM.
  - **CMS is the issuer** (creates tokens).
  - **CRM is verifier only** (validates token, enforces RBAC on `/admin/*`).

### Token storage (Console, MVP)
- Console stores token client-side (e.g., localStorage) and attaches Bearer token on every CMS/CRM request.
- Console must handle `401` by clearing token and redirecting to `/login`.

---

## Mandatory Implementation Stack (v1)

### Web framework
- **Gin** must be the HTTP framework (replace Gorilla Mux routing).
- Use Gin route groups and middleware for:
  - `request_id`
  - structured logging
  - panic recovery
  - CORS
  - JWT auth verification on `/admin/*`

### ORM + database
- **PostgreSQL** must be the system of record.
- **GORM** must be used for persistence.
- Migrations must be included (choose one):
  - `golang-migrate/migrate` (recommended), or
  - GORM AutoMigrate only for local/dev (not for prod).

### Packaging & deployment
- The service must be runnable via:
  - `docker build` + `docker run`
  - `docker compose up` (with Postgres)

---

## Purpose
The CRM Service is the canonical backend for managing customer relationships (customers/contacts/deals/activities) and supporting operational workflows (assignments, follow-ups, notes, attachments, basic reporting).

Platform Console is the **frontend/control surface** for CRM operations (customers/deals/activities) while CRM Service remains the system-of-record and rule enforcer.

---

## Current State (Repository)
The current repo is a simple Go service using a `net/http` router (documented as Gorilla Mux) with JSON-file-backed customers CRUD endpoints (`/customers` etc.), running on port `:3000`.

This requirements document defines what must be added/changed to make it production-ready and integratable with Platform Console.

---

## Scope

### In scope for v1 (Console-integratable CRM)
1) **Admin auth verification** compatible with Platform Console auth model, with a dedicated `/admin` route group and RBAC.
2) **Core CRM modules**: Customers, Contacts, Deals, Activities/Tasks, Notes, Tags.
3) **Operational visibility** endpoints (list/detail with filtering + pagination).
4) **Minimum “important” actions** (assignment, status/stage transitions, schedule follow-up).
5) **Non-functional production readiness**: Postgres persistence, migrations, observability, CI checks.

### Out of scope for v1
- End-user/customer portal UI.
- Advanced marketing automation (drip campaigns), unless required by your main product.
- Multi-tenant SaaS billing and tenant provisioning (future).

---

## Users & Permissions

### Roles (initial)
- **Admin**: Full access; can manage users, pipelines, configuration.
- **Manager**: Can view team performance, reassign work, manage deals.
- **Agent/SalesRep**: Can manage assigned customers/deals/activities.

### Authorization requirements
- Every `/admin/*` request must include `Authorization: Bearer <token>`.
- CRM must verify JWT with HS256 using `JWT_SECRET` and reject:
  - Missing/invalid token → `401`
  - Valid token but role not allowed → `403`
- Prefer consistent RBAC checks across all endpoints.

---

## Functional Requirements

## Admin Auth (Verifier mode)
> **Important:** For MVP, CRM does NOT issue tokens; CMS issues tokens and CRM verifies them.

### FR-Auth-1: Verify token + identity
- `GET /admin/me` must return current user identity + role/permissions derived from the JWT claims.
- JWT must include at minimum:
  - `sub` (user id) OR `user_id`
  - `role` (admin/manager/agent)
  - `exp` (expiry)

### FR-Auth-2: Logout behavior (MVP)
- CRM logout endpoint is optional.
- Console can “logout” by clearing token client-side and redirecting to `/login`.

### FR-Auth-3: RBAC enforcement
- All CRM `/admin/*` endpoints must enforce RBAC.
- Include an explicit permission model for:
  - Read vs write
  - “Own records” vs “all records” (future-friendly)

---

## Customers (Core)
### FR-Cust-1: Create customer
- Create a customer record with required identity fields.
- Must validate email format and uniqueness.

### FR-Cust-2: List customers
- List customers with server-side pagination.
- Filters: status, assigned_to, tags, created range, search (name/email/company).

### FR-Cust-3: Customer details view
- Return customer profile plus related entities summary:
  - Contacts, open deals, upcoming activities, recent activity timeline.

### FR-Cust-4: Update & partial update
- Support PUT and PATCH updates.
- PATCH should allow updating fields like `status`, `assigned_to`, `contacted`, `next_follow_up_at`.

### FR-Cust-5: Soft delete
- Support soft deletion (recoverable) to protect data integrity.

---

## Contacts
### FR-Contacts-1: Manage customer contacts
- CRUD contacts under a customer.
- Support primary contact designation.

---

## Deals / Opportunities
### FR-Deals-1: Deals CRUD
- CRUD deals with pipeline stage, amount, probability, expected close date.

### FR-Deals-2: Pipeline stages
- Stages configurable by admins.
- Deal stage transitions must be validated.

### FR-Deals-3: List & filter
- Filter by stage, owner, customer, date range, amount range.

---

## Activities / Tasks
### FR-Act-1: Activity tracking
- Log calls/emails/meetings/notes/tasks.
- Must support scheduled vs completed.

### FR-Act-2: Reminders
- Create activities with due dates.
- Provide “my tasks” endpoint for the current user.

---

## Notes & Comments
### FR-Notes-1: Notes CRUD
- Notes attached to customers and/or deals.
- Track author and timestamps.

---

## Tags & Categorization
### FR-Tags-1: Tags management
- CRUD tags.
- Assign/unassign tags to customers.

---

## Attachments (optional v1, recommended)
### FR-Attach-1: File upload
- Upload attachments for customer/deal/activity.
- Support S3-compatible storage (MinIO/S3) + metadata in Postgres.

---

## Audit Logging (recommended)
### FR-Audit-1: Immutable audit trail
- Record who changed what (resource, field diffs, timestamp).
- Expose read-only audit endpoints for admins.

---

## Reporting / Analytics (minimal v1)
### FR-Report-1: Basic dashboards
- Count customers by status.
- Count deals by stage.
- Revenue totals for won deals (if tracked).

---

## Required Backend Endpoints (Updated)

> These define what CRM must implement for Platform Console integration.

### Admin Auth (CRM = verifier)
- `GET /admin/me`

### Customers
- `GET /admin/customers` (filters + pagination)
- `POST /admin/customers`
- `GET /admin/customers/:id`
- `PUT /admin/customers/:id`
- `PATCH /admin/customers/:id`
- `DELETE /admin/customers/:id`

### Contacts
- `GET /admin/customers/:id/contacts`
- `POST /admin/customers/:id/contacts`
- `PUT /admin/contacts/:id`
- `DELETE /admin/contacts/:id`

### Deals
- `GET /admin/deals`
- `POST /admin/deals`
- `GET /admin/deals/:id`
- `PUT /admin/deals/:id`
- `PATCH /admin/deals/:id` (stage transitions)
- `DELETE /admin/deals/:id`

### Activities
- `GET /admin/activities` (filters + pagination)
- `POST /admin/activities`
- `GET /admin/activities/:id`
- `PUT /admin/activities/:id`
- `PATCH /admin/activities/:id` (status complete/cancel)
- `DELETE /admin/activities/:id`
- `GET /admin/me/activities` (my tasks)

### Tags
- `GET /admin/tags`
- `POST /admin/tags`
- `PUT /admin/tags/:id`
- `DELETE /admin/tags/:id`

### Reporting (minimal)
- `GET /admin/reports/overview`

### Health/metrics
- `GET /health`
- `GET /metrics`

---

## Security & CORS (Vercel cross-origin)

### CORS requirements
Because Platform Console runs on Vercel (different origin), CRM must:
- Allow cross-origin requests from Console origin (production + staging + localhost).
- Allow the `Authorization` header.
- Allow methods: `GET, POST, PUT, PATCH, DELETE, OPTIONS`.
- Apply CORS primarily to `/admin/*` (and optionally all routes).

### 401/403 contract
- Missing/invalid token → `401` with stable error shape.
- Not allowed role → `403` with stable error shape.

---

## Environment Variables (Updated)

### CRM (must match CMS)
- `JWT_SECRET` (HS256 shared secret used to verify tokens).
- `JWT_ISSUER` (optional, if you want to check issuer claim).
- `CORS_ALLOWED_ORIGINS` (comma-separated list; include Console’s Vercel domains).

### CRM (database)
- `DB_HOST`
- `DB_PORT`
- `DB_NAME`
- `DB_USER`
- `DB_PASSWORD`
- `DB_SSLMODE` (e.g., `disable` for local compose)

### Platform Console
- `NEXT_PUBLIC_CMS_BASE_URL`
- `NEXT_PUBLIC_CRM_BASE_URL`

---

## Docker Requirements

### Deliverables
- A production-ready `Dockerfile` (multi-stage build recommended).
- A `docker-compose.yml` that runs:
  - `postgres` (required)
  - `crm-service` (this repo)
  - optional `migrate` job/container to apply DB migrations

### Dockerfile requirements
- Expose the CRM HTTP port (e.g., `3000`).
- Read all config via environment variables.
- Container should be stateless (no local JSON file DB).

### docker-compose requirements (local dev)
- Must include Postgres volume.
- Must set CRM env vars (`DB_*`, `JWT_SECRET`, `CORS_ALLOWED_ORIGINS`).
- Must support `docker compose up` bringing the full stack up.

---

## Non-Functional Requirements

## Performance
- Default page size ~20.
- Server-side filtering/sorting.
- Prefer cursor pagination for large lists (future).

## Reliability
- Postgres as system of record (no JSON file DB).
- Background jobs for reminders (Redis queue) if needed.

## Observability
- Structured logs (JSON) with request IDs.
- Metrics endpoint for Prometheus.
- Tracing-ready middleware (optional).

## Deployment
- Container-first (Docker).
- Environment-based config.

---

## Acceptance Criteria (v1)
- Admin logs in once via CMS and Console can call both CMS and CRM using the same Bearer token.
- CRM rejects missing/invalid tokens with `401` and disallowed roles with `403` for `/admin/*`.
- Console handles `401` by clearing stored token and redirecting to `/login`.
- Admin can manage customers/deals/activities via Console against CRM APIs (pagination + filtering supported).
- CRM runs on Gin, persists to Postgres via GORM, and is runnable with Docker and docker-compose.
