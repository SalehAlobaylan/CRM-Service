# 🏢 CRM Service

A production-ready **Customer Relationship Management** API backend built with Go, designed for Platform Console integration.

## 🛠️ Built With

- **Go 1.24** - Backend language
- **Gin** - Web framework
- **GORM** - ORM for PostgreSQL
- **PostgreSQL** - Database
- **JWT (HS256)** - Authentication
- **Docker** - Containerization

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- PostgreSQL (or use Docker)

### Run with Docker Compose (Recommended)

```bash
# Copy environment file
cp .env.example .env

# Start all services (PostgreSQL + CRM)
docker compose up -d

# View logs
docker compose logs -f crm-service
```

### Run Locally

```bash
# Start PostgreSQL (via Docker or locally)
docker compose up -d postgres

# Copy and configure environment
cp .env.example .env

# Run the service
go run ./cmd/server
```

The service will be available at `http://localhost:3000`

## 📋 API Endpoints

### Health & Metrics

| Method | Endpoint   | Description        |
| ------ | ---------- | ------------------ |
| GET    | `/health`  | Health check       |
| GET    | `/ready`   | Readiness check    |
| GET    | `/metrics` | Prometheus metrics |

### Authentication (JWT Required)

| Method | Endpoint               | Description           |
| ------ | ---------------------- | --------------------- |
| GET    | `/admin/me`            | Get current user info |
| GET    | `/admin/me/activities` | Get my assigned tasks |

### Customers

| Method | Endpoint               | Description                |
| ------ | ---------------------- | -------------------------- |
| GET    | `/admin/customers`     | List customers (paginated) |
| POST   | `/admin/customers`     | Create customer            |
| GET    | `/admin/customers/:id` | Get customer details       |
| PUT    | `/admin/customers/:id` | Update customer            |
| PATCH  | `/admin/customers/:id` | Partial update             |
| DELETE | `/admin/customers/:id` | Soft delete customer       |

### Contacts

| Method | Endpoint                        | Description            |
| ------ | ------------------------------- | ---------------------- |
| GET    | `/admin/customers/:id/contacts` | List customer contacts |
| POST   | `/admin/customers/:id/contacts` | Create contact         |
| PUT    | `/admin/contacts/:id`           | Update contact         |
| DELETE | `/admin/contacts/:id`           | Delete contact         |

### Deals

| Method | Endpoint           | Description            |
| ------ | ------------------ | ---------------------- |
| GET    | `/admin/deals`     | List deals (paginated) |
| POST   | `/admin/deals`     | Create deal            |
| GET    | `/admin/deals/:id` | Get deal details       |
| PUT    | `/admin/deals/:id` | Update deal            |
| PATCH  | `/admin/deals/:id` | Stage transition       |
| DELETE | `/admin/deals/:id` | Soft delete deal       |

### Activities

| Method | Endpoint                | Description              |
| ------ | ----------------------- | ------------------------ |
| GET    | `/admin/activities`     | List activities          |
| POST   | `/admin/activities`     | Create activity          |
| GET    | `/admin/activities/:id` | Get activity details     |
| PUT    | `/admin/activities/:id` | Update activity          |
| PATCH  | `/admin/activities/:id` | Complete/cancel activity |
| DELETE | `/admin/activities/:id` | Delete activity          |

### Tags

| Method | Endpoint                           | Description             |
| ------ | ---------------------------------- | ----------------------- |
| GET    | `/admin/tags`                      | List all tags           |
| POST   | `/admin/tags`                      | Create tag (admin only) |
| PUT    | `/admin/tags/:id`                  | Update tag (admin only) |
| DELETE | `/admin/tags/:id`                  | Delete tag (admin only) |
| POST   | `/admin/customers/:id/tags/:tagId` | Assign tag              |
| DELETE | `/admin/customers/:id/tags/:tagId` | Remove tag              |

### Reports

| Method | Endpoint                  | Description        |
| ------ | ------------------------- | ------------------ |
| GET    | `/admin/reports/overview` | Dashboard overview |

## 🔐 Authentication

All `/admin/*` endpoints require JWT authentication via the `Authorization` header:

```
Authorization: Bearer <jwt_token>
```

JWT tokens are issued by the CMS service. CRM only verifies tokens using the shared `JWT_SECRET`.

### JWT Claims Required

```json
{
  "user_id": 123,
  "role": "admin|manager|agent",
  "exp": 1234567890
}
```

### Roles & Permissions

| Role    | Permissions                     |
| ------- | ------------------------------- |
| admin   | Full access, manage tags/config |
| manager | Manage all records              |
| agent   | Manage own assigned records     |

## ⚙️ Environment Variables

| Variable               | Description                   | Default                 |
| ---------------------- | ----------------------------- | ----------------------- |
| `SERVER_PORT`          | HTTP port                     | `3000`                  |
| `ENVIRONMENT`          | `development` or `production` | `development`           |
| `DB_HOST`              | PostgreSQL host               | `localhost`             |
| `DB_PORT`              | PostgreSQL port               | `5432`                  |
| `DB_NAME`              | Database name                 | `crm_db`                |
| `DB_USER`              | Database user                 | `postgres`              |
| `DB_PASSWORD`          | Database password             | `postgres`              |
| `DB_SSLMODE`           | SSL mode                      | `disable`               |
| `JWT_SECRET`           | HS256 signing key             | (required)              |
| `JWT_ISSUER`           | Expected issuer               | `cms`                   |
| `CORS_ALLOWED_ORIGINS` | Comma-separated origins       | `http://localhost:3000` |

## 📁 Project Structure

```
CRM-Service/
├── cmd/
│   └── server/
│       └── main.go           # Entry point
├── internal/
│   ├── config/               # Environment configuration
│   ├── database/             # GORM + PostgreSQL setup
│   ├── middleware/           # JWT auth, CORS, logging
│   ├── models/               # GORM models
│   ├── handlers/             # HTTP handlers
│   └── routes/               # Gin router setup
├── migrations/               # SQL migrations
├── dockerfile                # Multi-stage Docker build
├── docker-compose.yml        # Local dev stack
└── .env.example              # Environment template
```

## 🐳 Docker Commands

```bash
# Build and start
docker compose up -d --build

# View logs
docker compose logs -f

# Stop services
docker compose down

# Reset database
docker compose down -v
docker compose up -d
```

## 📊 Query Examples

### List Customers with Filters

```bash
curl "http://localhost:3000/admin/customers?status=active&page=1&page_size=20&search=john" \
  -H "Authorization: Bearer <token>"
```

### Create a Deal

```bash
curl -X POST "http://localhost:3000/admin/deals" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Enterprise License",
    "customer_id": 1,
    "amount": 50000,
    "stage": "qualification"
  }'
```

## 📝 License

MIT

### 🔄 Update Contacted Status Only

```
PATCH /customers/{id}
```

### ❌ Delete Customer

```
DELETE /customers/{id}
```

## 🧪 Test the API

```bash
# Get all customers
curl http://localhost:3000/customers

# Create a customer
curl -X POST http://localhost:3000/customers \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","role":"Manager","email":"john@example.com","phone":"123-456-7890","contacted":false}'
```

## 🌐 Server

Server runs on **http://localhost:3000**
