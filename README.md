# 🏢 CRM Service

A simple **Customer Relationship Management** service built with Go.
# https://crm-system-seven-ecru.vercel.app/
## 🛠️ Built With

- **Go** - Backend language
- **Gorilla Mux** - HTTP router

### Run with Go

```bash
go run main.go
```

### Run with Docker

```bash
# Build the image
docker build -t crm-system .

# Run the container
docker run -p 3000:3000 crm-system
```

## 📋 API Endpoints

### 👥 Get All Customers

```
GET /customers
```

### 👤 Get Customer by ID

```
GET /customers/{id}
```

### ➕ Create New Customer

```
POST /customers
```

### ♻️ Update Customer (Full Replace)

```
PUT /customers/{id}
```

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
