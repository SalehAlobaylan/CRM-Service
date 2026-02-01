# CRM Service

A production-ready, API-first CRM backend for Turfa Platform, integrated with shared `turfa_platform` database.

## Overview

The CRM Service is a Go-based backend that provides customer relationship management functionality for Platform Console. It uses PostgreSQL for data persistence, GORM for ORM, and Gin for HTTP routing.

### Key Features

- **Full CRM Functionality**: Customers, Contacts, Deals, Activities, Tags
- **JWT Authentication**: Verifies CMS-issued JWT tokens (shared secret)
- **RBAC**: Role-based access control (Admin, Manager, Agent)
- **Soft Deletes**: All records support soft delete for data integrity
- **Pagination & Filtering**: Efficient querying with server-side filtering
- **Audit Logging**: Immutable record of all changes
- **Metrics & Observability**: Prometheus metrics, structured logging

## Architecture

### Shared Database

The CRM Service now shares `turfa_platform` database with CMS Service:

| Service | Tables | Primary Key Type | Soft Delete |
|---------|--------|------------------|-------------|
| **CMS** | `blogs`, `categories`, `content_items`, `content_sources`, `media`, `pages`, `posts`, `transcripts`, `user_interactions`, `visitors` | `uuid` | No |
| **CRM** | `customers`, `contacts`, `pipeline_stages`, `deals`, `activities`, `notes`, `tags`, `customer_tags`, `audit_logs` | `SERIAL` | Yes |

**Conflict Status:** No conflicts - all table names are unique across services.

### Authentication Model

- **CMS is the issuer** - creates and signs JWT tokens
- **CRM is verifier only** - validates tokens and enforces RBAC
- **Shared JWT_SECRET** - Both services must use the same secret
- **Authorization Header** - `Authorization: Bearer <token>` on every request

## Implementation Status

| Module                     | Status         | Notes                                          |
| -------------------------- | -------------- | ---------------------------------------------- |
| Gin Framework              | ✅ Complete    | Replaced Gorilla Mux with Gin                  |
| PostgreSQL + GORM          | ✅ Complete    | Full persistence layer                         |
| JWT Auth (HS256 Verifier)  | ✅ Complete    | Middleware validates CMS-issued tokens         |
| RBAC (admin/manager/agent) | ✅ Complete    | Role-based access control                      |
| CORS Middleware            | ✅ Complete    | Configured for Vercel origins                  |
| Customers CRUD             | ✅ Complete    | With pagination, filtering, soft delete        |
| Contacts CRUD              | ✅ Complete    | Nested under customers, primary designation    |
| Deals CRUD                 | ✅ Complete    | Pipeline stages, transitions                   |
| Activities CRUD            | ✅ Complete    | Including `/admin/me/activities`               |
| Tags CRUD                  | ✅ Complete    | With customer assignment                       |
| Reports Overview           | ✅ Complete    | `/admin/reports/overview`                      |
| Health/Ready/Metrics       | ✅ Complete    | Prometheus metrics                             |
| Structured Logging         | ✅ Complete    | Zap JSON logs with request IDs                 |
| Docker + Compose           | ✅ Complete    | Multi-stage build, PostgreSQL                  |
| SQL Migrations             | ✅ Complete    | golang-migrate compatible                      |
| **Notes CRUD**             | ⚠️ Partial     | Model exists, **endpoints not implemented**    |
| **Audit Read Endpoint**    | ⚠️ Partial     | Logs written, **GET endpoint not implemented** |
| Attachments/File Upload    | ❌ Not Started | Optional for v1                                |

## Quick Start

### Prerequisites

- Go 1.24 or higher
- PostgreSQL 15 or higher
- golang-migrate (optional, for production migrations)

### Installation

```bash
# Clone repository
git clone https://github.com/SalehAlobaylan/CRM-Service.git
cd CRM-Service

# Install dependencies
go mod download
```

### Configuration

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env` with your settings:

```env
# Shared Database (Turfa Platform)
DATABASE_URL=postgres://postgres:your-password@your-postgres-host:5432/turfa_platform?sslmode=disable

# Shared JWT (Must match CMS)
JWT_SECRET=your-shared-secret-key
JWT_ISSUER=cms

# CORS (Include Platform Console origins)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:3001,https://your-console.vercel.app
```

### Database Migrations

The CRM Service uses `golang-migrate` for database migrations in production.

#### Using golang-migrate CLI

```bash
# Install golang-migrate
# MacOS:   brew install golang-migrate
# Linux:   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz

# Run migrations (up)
migrate -path ./migrations -database "${DATABASE_URL}" up

# Rollback last migration
migrate -path ./migrations -database "${DATABASE_URL}" down 1

# Show current version
migrate -path ./migrations -database "${DATABASE_URL}" version

# Force migration to version N
migrate -path ./migrations -database "${DATABASE_URL}" force N

# Create new migration
migrate create -ext sql -dir ./migrations -seq add_new_feature
```

#### Using Docker Compose (optional)

```bash
# Run migrations via docker-compose
docker-compose --profile migrate up migrate
```

### Running Service

#### Local Development

```bash
# Run directly
go run src/main.go
```

#### Using Docker

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Health Check

```bash
curl http://localhost:3000/health
```

## API Documentation

### Base URL

```
http://localhost:3000
```

### Public Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/ready` | Readiness probe |
| GET | `/metrics` | Prometheus metrics |

### Admin Endpoints (JWT Required)

All admin endpoints require `Authorization: Bearer <token>` header.

#### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/me` | Get current user info |
| GET | `/admin/me/activities` | Get my activities |

#### Customers

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/customers` | List customers (with pagination) |
| POST | `/admin/customers` | Create customer |
| GET | `/admin/customers/:id` | Get customer details |
| PUT | `/admin/customers/:id` | Update customer |
| PATCH | `/admin/customers/:id` | Partial update customer |
| DELETE | `/admin/customers/:id` | Soft delete customer |
| GET | `/admin/customers/:id/contacts` | List customer contacts |
| POST | `/admin/customers/:id/contacts` | Add contact to customer |
| POST | `/admin/customers/:id/tags/:tagId` | Assign tag to customer |
| DELETE | `/admin/customers/:id/tags/:tagId` | Remove tag from customer |

#### Contacts

| Method | Endpoint | Description |
|--------|----------|-------------|
| PUT | `/admin/contacts/:id` | Update contact |
| DELETE | `/admin/contacts/:id` | Delete contact |

#### Deals

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/deals` | List deals |
| POST | `/admin/deals` | Create deal |
| GET | `/admin/deals/:id` | Get deal details |
| PUT | `/admin/deals/:id` | Update deal |
| PATCH | `/admin/deals/:id` | Partial update deal |
| DELETE | `/admin/deals/:id` | Delete deal |

#### Activities

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/activities` | List activities |
| POST | `/admin/activities` | Create activity |
| GET | `/admin/activities/:id` | Get activity details |
| PUT | `/admin/activities/:id` | Update activity |
| PATCH | `/admin/activities/:id` | Partial update activity |
| DELETE | `/admin/activities/:id` | Delete activity |

#### Tags

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/tags` | List tags |
| POST | `/admin/tags` | Create tag (Admin only) |
| PUT | `/admin/tags/:id` | Update tag (Admin only) |
| DELETE | `/admin/tags/:id` | Delete tag (Admin only) |

#### Reports

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/admin/reports/overview` | Get overview report |

## Project Structure

```
CRM-Service/
├── src/
│   ├── main.go                 # Application entry point
│   ├── config/                 # Configuration loading
│   ├── database/               # Database connection
│   ├── handlers/               # HTTP request handlers
│   ├── middleware/             # Custom middleware (auth, CORS, logging)
│   ├── models/                 # Data models
│   └── routes/                 # Route definitions
├── migrations/                 # SQL migrations
├── context/                    # Context documentation
├── docker-compose.yml          # Docker Compose configuration
├── dockerfile                  # Docker image build
├── go.mod                      # Go module dependencies
└── .env                        # Environment variables
```

## Database Integration

For detailed information about database integration with Turfa Platform, see:
- [DATABASE_INTEGRATION_GUIDE.md](./DATABASE_INTEGRATION_GUIDE.md)

## Security Considerations

1. **JWT Secret**: Always use a strong, randomly generated JWT_SECRET in production. Ensure it matches CMS service's JWT_SECRET.
2. **Database URL**: Never commit actual database credentials. Use environment variables.
3. **CORS**: Configure CORS_ALLOWED_ORIGINS to only include trusted origins.
4. **SSL**: Enable SSL mode (`sslmode=require` or `sslmode=verify-ca`) in production database connections.

## Production Deployment

### Environment Variables

Required environment variables for production:

```env
SERVER_PORT=3000
ENVIRONMENT=production
DATABASE_URL=postgres://user:pass@host:5432/turfa_platform?sslmode=require
JWT_SECRET=<strong-random-secret>
JWT_ISSUER=cms
CORS_ALLOWED_ORIGINS=https://console.yourdomain.com
```

### Docker Deployment

```bash
# Build production image
docker build -t crm-service:prod .

# Run with production configuration
docker run -d \
  --name crm-service \
  -p 3000:3000 \
  -e SERVER_PORT=3000 \
  -e ENVIRONMENT=production \
  -e DATABASE_URL="postgres://..." \
  -e JWT_SECRET="..." \
  -e CORS_ALLOWED_ORIGINS="https://console.yourdomain.com" \
  crm-service:prod
```

## Troubleshooting

### Database Connection Issues

```bash
# Test database connection
psql "postgres://user:pass@host:5432/turfa_platform?sslmode=disable"

# Check migration status
migrate -path ./migrations -database "${DATABASE_URL}" version
```

### JWT Authentication Issues

Ensure:
1. JWT_SECRET is identical between CMS and CRM services
2. Token includes required claims: `role`, `exp`, and either `sub` or `user_id`
3. Authorization header is formatted correctly: `Bearer <token>`

### CORS Issues

Verify:
1. CORS_ALLOWED_ORIGINS includes the requesting origin
2. OPTIONS requests are allowed (handled by CORS middleware)

## Related Services

- **CMS Service**: https://github.com/SalehAlobaylan/CMS-Service
- **Platform Console**: https://github.com/SalehAlobaylan/Platform-Console
- **Aggregation Service**: https://github.com/SalehAlobaylan/Aggregation-Service

## Context Documentation

- [CRM_Context_Requirements.md](./context/CRM_Context_Requirements.md)
- [Platform_Console_Context_Requirements.md](./context/Platform_Console_Context_Requirements.md)
- [DATABASE_INTEGRATION_GUIDE.md](./DATABASE_INTEGRATION_GUIDE.md)
- [Refactoring_Summary.md](./context/Refactoring_Summary.md)
