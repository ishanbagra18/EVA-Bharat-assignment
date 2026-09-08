# Ticket System REST API Backend

A production-grade REST API backend written in Go for managing support tickets with ownership-based authorization, clean state machine status transitions, and SQLite persistence.

---

## Tech Stack

- **Language**: Go (`1.24+`)
- **HTTP Router**: `github.com/go-chi/chi/v5`
- **Authentication**: JWT (`github.com/golang-jwt/jwt/v5`) & Bcrypt password hashing (`golang.org/x/crypto/bcrypt`)
- **Database**: SQLite via `modernc.org/sqlite` (Pure Go driver, `CGO_ENABLED=0` friendly)
- **ID Generation**: UUID v4 (`github.com/google/uuid`)
- **Containerization**: Multi-stage Docker build (`alpine:latest`)

---

## Architecture & Project Structure

```
ticket-system/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point & dependency injection
├── internal/
│   ├── auth/
│   │   ├── jwt.go               # Token generation & verification logic
│   │   ├── password.go          # Bcrypt hashing utilities
│   │   └── middleware.go        # JWT authentication middleware
│   ├── db/
│   │   └── db.go                # SQLite database initialization & migrations
│   ├── handlers/
│   │   ├── auth_handler.go      # POST /auth/register, POST /auth/login
│   │   ├── ticket_handler.go    # Ticket CRUD & status endpoints
│   │   ├── health_handler.go    # GET /health
│   │   └── response.go          # Standardized JSON response helpers
│   ├── models/
│   │   ├── user.go              # User models & DTOs
│   │   └── ticket.go            # Ticket models & status enums
│   ├── repository/
│   │   ├── user_repo.go         # UserRepository interface & SQLite implementation
│   │   └── ticket_repo.go       # TicketRepository interface & SQLite implementation
│   ├── router/
│   │   └── router.go            # Chi router setup & route registration
│   └── service/
│       ├── auth_service.go      # Registration & Login business logic
│       ├── ticket_service.go    # Ownership authorization & state machine logic
│       └── ticket_service_test.go # State machine unit tests
├── migrations/
│   └── 001_init.sql             # Database SQL schema
├── Dockerfile                   # Multi-stage Docker packaging
├── .env.example                 # Example environment variables
├── .gitignore                   # Ignore rules
├── go.mod / go.sum              # Go module definitions
└── README.md                    # Project documentation
```

---

## Getting Started

### 1. Environment Configuration

Copy the example environment file:
```bash
cp .env.example .env
```

Default configuration (`.env`):
```env
PORT=8080
JWT_SECRET=super_secret_jwt_key_replace_in_production
JWT_EXPIRY_HOURS=24
DB_PATH=./data/tickets.db
```

### 2. Local Run (Go CLI)

```bash
# Run tests
go test ./... -v

# Start server
go run ./cmd/server/main.go
```

The server will initialize the SQLite database at `./data/tickets.db` and listen on `http://localhost:8080`.

### 3. Local Run (Docker)

```bash
# Build image
docker build -t ticket-system .

# Run container
docker run -p 8080:8080 --env-file .env ticket-system
```

---

## API Reference

### Health Check

#### `GET /health`
Returns system status.
- **Auth**: None
- **Response**: `200 OK`
```json
{
  "status": "ok"
}
```

---

### Authentication

#### `POST /auth/register`
Register a new user account.
- **Auth**: None
- **Request**:
```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```
- **Response (`201 Created`)**:
```json
{
  "id": "11111111-2222-3333-4444-555555555555",
  "email": "user@example.com",
  "created_at": "2026-09-08T22:00:00Z"
}
```

#### `POST /auth/login`
Authenticate user and obtain a JWT bearer token.
- **Auth**: None
- **Request**:
```json
{
  "email": "user@example.com",
  "password": "secretpassword"
}
```
- **Response (`200 OK`)**:
```json
{
  "token": "eyJhbGciOiJIUzI1Ni..."
}
```

---

### Ticket Management (Auth Required)

*Note: All `/tickets` endpoints require the header `Authorization: Bearer <token>`.*

#### `POST /tickets`
Create a new ticket owned by the authenticated user.
- **Request**:
```json
{
  "title": "Application fails on login",
  "description": "Getting error code 500 when clicking login button."
}
```
- **Response (`201 Created`)**:
```json
{
  "id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "user_id": "11111111-2222-3333-4444-555555555555",
  "title": "Application fails on login",
  "description": "Getting error code 500 when clicking login button.",
  "status": "open",
  "created_at": "2026-09-08T22:05:00Z",
  "updated_at": "2026-09-08T22:05:00Z"
}
```

#### `GET /tickets`
List all tickets created by the authenticated user.
- **Query Params**: `?status=open` (optional filter: `open`, `in_progress`, `closed`)
- **Response (`200 OK`)**:
```json
[
  {
    "id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "user_id": "11111111-2222-3333-4444-555555555555",
    "title": "Application fails on login",
    "description": "Getting error code 500 when clicking login button.",
    "status": "open",
    "created_at": "2026-09-08T22:05:00Z",
    "updated_at": "2026-09-08T22:05:00Z"
  }
]
```

#### `GET /tickets/{id}`
Retrieve a specific ticket.
- **Responses**:
  - `200 OK`: If the ticket exists and belongs to the authenticated user.
  - `403 Forbidden`: If the ticket exists but belongs to a different user (`{"error": "access denied: resource belongs to another user"}`).
  - `404 Not Found`: If the ticket ID does not exist (`{"error": "ticket not found"}`).

#### `PATCH /tickets/{id}/status`
Update ticket status according to state machine rules.
- **Request**:
```json
{
  "status": "in_progress"
}
```
- **State Machine Rules**:
  - `open` -> `in_progress` ✅
  - `open` -> `closed` ✅
  - `in_progress` -> `closed` ✅
  - Same status (e.g. `open` -> `open`) ✅ (idempotent, returns `200 OK`)
  - `closed` -> `open` or `closed` -> `in_progress` ❌ (`409 Conflict`: Closed state is terminal)
- **Responses**:
  - `200 OK`: Updated ticket object.
  - `400 Bad Request`: Invalid status value.
  - `403 Forbidden`: Authenticated user does not own the ticket.
  - `404 Not Found`: Ticket ID does not exist.
  - `409 Conflict`: Illegal status transition (e.g., trying to reopen a closed ticket).

---

## Standardized Error Response Format

All error responses return a uniform JSON payload structure:
```json
{
  "error": "human readable error message"
}
```

---

## Architectural Assumptions

1. **Ownership Authorization (403 vs 404)**: Accessing a ticket owned by another user yields `403 Forbidden` rather than `404 Not Found` to explicitly communicate an authorization failure.
2. **Terminal State**: `closed` status is terminal. Reopening closed tickets is prohibited and yields `409 Conflict`.
3. **Pure-Go SQLite Persistence**: Uses `modernc.org/sqlite` driver to allow zero-CGO compilation (`CGO_ENABLED=0`) for lightweight, cross-platform Docker deployments.
4. **JWT Expiry**: JWTs expire in 24 hours (configurable via `JWT_EXPIRY_HOURS`).
