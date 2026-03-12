# DevOps-first

A lightweight local CI/CD backend built with Go, Gin, and Gorilla WebSocket with **JWT authentication**, MySQL database, and a Vue3 + Ant Design Vue + Xterm frontend for real-time deployment logs.

## Project Structure

```text
.
├── cmd
│   └── server
│       └── main.go
├── internal
│   ├── database
│   │   └── db.go
│   ├── handler
│   │   ├── auth.go
│   │   └── deploy.go
│   ├── middleware
│   │   └── jwt.go
│   ├── model
│   │   └── user.go
│   └── service
│       ├── auth.go
│       └── executor.go
├── web
│   ├── src
│   └── .env.example
├── .env.example
├── go.mod
└── PROGRESS.md
```

## Prerequisites

- Go 1.22+
- Node.js 20+
- Docker
- Maven
- Git
- **MySQL 5.7+**

## Environment Variables

Backend (`.env.example`):

- `HTTP_ADDR` default `:8081`
- `GIN_MODE` default `debug` (use `release` in production)
- `TRUSTED_PROXIES` default `127.0.0.1,::1`
- `GIT_PATH` default `git`
- `MVN_PATH` default `mvn`
- `DOCKER_PATH` default `docker`
- **`DB_HOST`** default `localhost`
- **`DB_PORT`** default `3306`
- **`DB_USER`** default `root`
- **`DB_PASSWORD`** default `` (empty)
- **`DB_NAME`** default `devops_first`

Frontend (`web/.env.example`):

- `VITE_BACKEND_TARGET` default `http://localhost:8081`

## Quick Start

### 0. One-command service lifecycle (recommended)

```bash
cd /Users/moryzang/GoProjects/DevOps-first

# Ensure backend/frontend are running and healthy
./scripts/dev-health-check.sh --ensure

# Check-only mode (returns non-zero if any service is down)
./scripts/dev-health-check.sh

# Stop backend/frontend
./scripts/dev-stop.sh
```

Logs and PID files are written to `.dev-runtime/`.

### 1. Database Setup

```bash
# Create MySQL database (user: root, password: silenceopr@2026)
mysql -u root -p'silenceopr@2026' -e "CREATE DATABASE IF NOT EXISTS devops_first CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# Tables are automatically created by GORM on first run
```

### 2. Backend

```bash
cd /Users/moryzang/GoProjects/DevOps-first
cp .env.example .env
set -a; source .env; set +a
go run ./cmd/server
```

Production-oriented local start example:

```bash
GIN_MODE=release \
DB_HOST=localhost \
DB_PORT=3306 \
DB_USER=root \
DB_PASSWORD=silenceopr@2026 \
DB_NAME=devops_first \
go run ./cmd/server
```

### 3. Frontend

```bash
cd /Users/moryzang/GoProjects/DevOps-first/web
cp .env.example .env
npm install
npm run dev
```

Open `http://localhost:5173`.

## Authentication

### API Endpoints

#### Register
```
POST /auth/register
Content-Type: application/json

{
  "username": "myuser",
  "password": "mypassword123",
  "email": "user@example.com",  // optional
  "remark": "my remark"         // optional
}

Response:
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-03-18T20:02:34+08:00",
  "user": {
    "id": 1,
    "username": "myuser",
    "email": "user@example.com",
    "remark": "my remark"
  }
}
```

#### Login
```
POST /auth/login
Content-Type: application/json

{
  "username": "myuser",
  "password": "mypassword123"
}

Response: (same as register)
```

### Frontend Flow

1. On first visit, user sees **Login/Register** form
2. After authentication, JWT `token` is saved to `localStorage`
3. Token is automatically passed to WebSocket via query parameter: `?token=...`
4. User can deploy projects via Xterm interface
5. Logout clears token and redirects to login form

### Token Management

- Token expiration: **7 days**
- Stored in browser `localStorage` for session persistence
- JWT Secret: `your-secret-key` (change in production via `internal/service/auth.go`)

## Usage

1. **Register or Login** on the frontend
2. Enter `project_path` in the deployment form
3. Click `Start Deploy`
4. Backend executes CI/CD pipeline:
   - `git pull`
   - `mvn clean package`
   - `docker build`
   - `docker run`
5. Logs stream line-by-line to Xterm panel in real time
6. Click `Stop` to cancel deployment

## Port Conflict Note

If `:8081` is in use, run backend on another port:

```bash
HTTP_ADDR=:8090 \
DB_HOST=localhost \
DB_PORT=3306 \
DB_USER=root \
DB_PASSWORD=silenceopr@2026 \
go run ./cmd/server
```

Then update frontend target for the current shell:

```bash
cd /Users/moryzang/GoProjects/DevOps-first/web
VITE_BACKEND_TARGET=http://localhost:8090 npm run dev
```

## Health Check

```bash
curl http://localhost:8081/healthz
```

Expected response:

```json
{"status":"ok"}
```

Or use the built-in script for both backend and frontend:

```bash
./scripts/dev-health-check.sh
```
