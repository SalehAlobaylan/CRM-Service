# CRM and CMS Database Integration Guide

This document outlines the strategy and technical details for merging the CRM Service logic into the shared Turfa Platform database.

## 1. Architecture Overview
The CRM Service and CMS Service now share a single PostgreSQL database instance (`turfa_platform`). 

- **CMS Tables**: Use `uuid` for primary keys and public IDs.
- **CRM Tables**: Use `SERIAL` (integer) for primary keys and include `deleted_at` for soft deletes.
- **Shared Authentication**: Both services verify JWTs using a shared `JWT_SECRET`.

---

## 2. Table Inventory & Conflict Check
All current CRM tables have been verified to have unique names that do not conflict with the existing CMS schema.

| Service | Tables |
| :--- | :--- |
| **CMS** | `blogs`, `categories`, `content_items`, `content_sources`, `media`, `pages`, `post_media`, `posts`, `transcripts`, `user_interactions`, `visitors` |
| **CRM** | `customers`, `contacts`, `pipeline_stages`, `deals`, `activities`, `notes`, `tags`, `customer_tags`, `audit_logs` |

**Conflict Status:** ✅ Zero conflicts found.

---

## 3. CRM Table Definitions
The following tables are implemented via the `migrations/000001_init_schema.up.sql` file:

### Core CRM Entities
*   **`customers`**: Primary identity for clients/leads.
*   **`contacts`**: Individual people nested under customers.
*   **`pipeline_stages`**: Configurable stages for the sales funnel (e.g., Prospecting, Closed Won).
*   **`deals`**: Sales opportunities linked to customers and stages.
*   **`activities`**: Tasks, calls, and meetings linked to customers/deals.
*   **`notes`**: Internal comments and logs for customer/deal history.
*   **`tags`**: Metadata labels for categorization.

### Junctions & Logs
*   **`customer_tags`**: Many-to-many relationship between customers and tags.
*   **`audit_logs`**: Immutable record of all changes made to CRM resources.

---

## 4. Integration Steps

### Step 1: Environment Configuration
Update your `.env` file to point to the shared platform database.

```env
# Shared Database
DB_HOST=your-shared-postgres-host
DB_PORT=5432
DB_NAME=turfa_platform
DB_USER=postgres
DB_PASSWORD=your-password
DB_SSLMODE=disable

# Authentication (Must match CMS)
JWT_SECRET=your-shared-secret-key
JWT_ISSUER=cms
```

### Step 2: Run Migrations
Apply the CRM schema to the existing CMS database using the `golang-migrate` tool:

```bash
migrate -database "postgres://user:password@host:5432/turfa_platform?sslmode=disable" -path ./migrations up
```

---

## 5. Security & RBAC
- **Authentication**: CRM acts as a **Verifier Only**. It does not issue tokens.
- **Middleware**: The Gin middleware checks `Authorization: Bearer <token>` against the shared secret.
- **Roles**: Claims must include `role` (Admin, Manager, Agent) to pass RBAC checks on `/admin/*` routes.

---

## 6. Maintenance
- **Backups**: Since all data is now in one database, a single backup of `turfa_platform` covers both services.
- **Cleanup**: CRM leverages **Soft Deletes**. Records are not physically removed unless the database is manually purged.
