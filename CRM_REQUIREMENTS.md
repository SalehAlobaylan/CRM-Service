# CRM Service Requirements (for Platform Console Integration)

**Target repo:** [SalehAlobaylan/CRM-Service](https://github.com/SalehAlobaylan/CRM-Service) [mcp_tool_github-mcp-direct_get_file_contents:0]

**Goal:** Refactor and scale the CRM service into a production-ready, API-first CRM backend that can be cleanly integrated into the existing **Platform Console** (admin dashboard) as a module, using HTTPS APIs and admin authentication. [mcp_tool_github-mcp-direct_get_file_contents:0]

---

## Purpose
The CRM Service is the canonical backend for managing customer relationships (customers/contacts/deals/activities) and supporting operational workflows (assignments, follow-ups, notes, attachments, basic reporting). [mcp_tool_github-mcp-direct_get_file_contents:0]

The Platform Console will act as the **frontend/control surface** for CRM operations (similar to how it manages `contentsources` and observes `contentitems.status` in the platform). [mcp_tool_github-mcp-direct_get_file_contents:0]

---

## Current State (Repository)
The current repo is a simple Go service using Gorilla Mux with JSON-file-backed customers CRUD endpoints (`/customers` etc.), running on port `:3000`. [mcp_tool_github-mcp-direct_get_file_contents:0][mcp_tool_github-mcp-direct_get_file_contents:1]

This requirements document defines what must be added/changed to make it production-ready and integratable with Platform Console. [mcp_tool_github-mcp-direct_get_file_contents:0]

---

## Scope

### In scope for v1 (Console-integratable CRM)
1) **Admin authentication** compatible with Platform Console auth model (JWT/SSO), with a dedicated `/admin` route group.  
2) **Core CRM modules**: Customers, Contacts, Deals, Activities/Tasks, Notes, Tags.  
3) **Operational visibility** endpoints (list/detail with filtering, pagination, audit history metadata).  
4) **Minimum “important” actions** (assignment, status/stage transitions, run automations like “schedule follow-up”).  
5) **Non-functional production readiness**: Postgres persistence, migrations, observability, CI checks.

### Out of scope for v1
- End-user/customer portal UI.
- Advanced marketing automation (drip campaigns), unless required by your main product.
- Multi-tenant SaaS billing and tenant provisioning (can be future).

---

## Users & Permissions

### Roles (initial)
- **Admin**: Full access; can manage users, pipelines, configuration.
- **Manager**: Can view team performance, reassign work, manage deals.
- **Agent/SalesRep**: Can manage assigned customers/deals/activities.

### Authorization requirements
- Every request must carry auth credentials (JWT or cookie session).
- Enforce permissions on the server side (RBAC) and return clear 403/401 errors.
- Prefer a dedicated route group: `/admin/*` for console use.

---

## Functional Requirements

## Admin Auth
### FR-Auth-1: Login / token exchange
- Provide a login/token exchange flow compatible with Platform Console (CMS-issued JWT/SSO style), even if CRM is a separate service.  
- Support short-lived access token + refresh token (or cookie-based session).

### FR-Auth-2: Session management
- Logout support (token invalidation if supported; otherwise revoke refresh token).
- `/admin/me` endpoint for “who am I” and role/permissions.

### FR-Auth-3: Service-to-service trust (optional)
- Allow Platform Console BFF to call CRM using an internal service credential (mTLS or signed service JWT).

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
- Provide “my tasks” endpoint for a given user.

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

## Required Backend Endpoints (Proposed)

> These define what CRM must implement for Platform Console integration.

### Admin Auth
- `POST /admin/auth/login`
- `POST /admin/auth/logout`
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

## Non-Functional Requirements

## Security
- Admin-only access for `/admin/*` endpoints.
- HTTPS only; secure cookies if using cookie auth.
- CORS configured for Platform Console origin.

## Performance
- Default page size ~20.
- Server-side filtering/sorting.
- Prefer cursor pagination for large lists (future).

## Reliability
- Postgres as system of record (no JSON file DB).
- Background jobs for notifications/reminders (Redis queue) if needed.

## Observability
- Structured logs (JSON) with request IDs.
- Metrics endpoint for Prometheus.
- Tracing-ready middleware (optional).

## Deployment
- Container-first (Docker).
- Environment-based config.

---

## Integration Notes (Platform Console)

### Recommended Integration
- Platform Console should call CRM via `/admin/*` endpoints over HTTPS.
- Prefer a BFF layer (Fastify) in Platform Console to:
  - proxy requests
  - handle cookies securely
  - reduce CORS complexity

### Environment variables
- `CRM_BASE_URL` (internal)
- `NEXT_PUBLIC_CRM_BASE_URL` (public)

---

## Acceptance Criteria (v1)
- Admin can authenticate and access customer list via Platform Console.  
- Admin can create/update a customer, assign it to a user, and see it in list/detail.  
- Admin can create a deal and move it through stages.  
- Admin can create activities with due dates and fetch “my tasks”.  
- API supports pagination + filtering for list pages.  
- Service runs against Postgres with migrations and is deployable with Docker.
