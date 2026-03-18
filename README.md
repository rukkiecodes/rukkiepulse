# RukkiePulse

A CLI-first observability and diagnostics platform for backend services.

Monitor health, probe endpoints, trace errors to their root cause — all from the terminal.

---

## Installation

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) (for Jaeger + OTel Collector — Phase 4+)

### Build from source

```bash
git clone https://github.com/rukkiecodes/rukkiepulse
cd rukkiepulse
go build -o rukkie ./cmd/rukkie/
```

Move the binary to your PATH:

```bash
mv rukkie /usr/local/bin/rukkie        # macOS / Linux
# Windows: move rukkie.exe to a folder in your PATH
```

---

## Quick Start

**1. Create a `rukkie.yaml` in your project root:**

```yaml
project: my-backend

observability:
  jaeger:
    url: http://localhost:16686

environments:
  dev:
    services:
      - name: auth-service
        url: http://localhost:3000
        type: REST
        endpoints:
          - path: /health
            method: GET
            expect_status: 200
          - path: /login
            method: POST
            body: '{"email":"test@test.com","password":"test"}'
            expect_status: 200

      - name: graphql-api
        url: http://localhost:4000/graphql
        type: GRAPHQL
        endpoints:
          - query: '{ __typename }'
            expect_no_errors: true
```

**2. Log in:**

```bash
rukkie login
```

**3. Scan your services:**

```bash
rukkie scan
```

---

## Configuration (`rukkie.yaml`)

| Field | Description |
|-------|-------------|
| `project` | Display name for the project |
| `observability.jaeger.url` | Jaeger Query UI URL (e.g. `http://localhost:16686`) |
| `environments.<name>.services` | List of services per environment |

### Service fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Display name |
| `url` | Yes | Base URL of the service |
| `type` | Yes | `REST` or `GRAPHQL` |
| `api_key` | No | API key for agent identity |
| `endpoints` | No | List of endpoints to probe |

### Endpoint fields (REST)

| Field | Default | Description |
|-------|---------|-------------|
| `path` | — | URL path e.g. `/login` |
| `method` | `GET` | HTTP method |
| `body` | — | JSON request body |
| `expect_status` | `200` | Expected HTTP status code |

### Endpoint fields (GraphQL)

| Field | Default | Description |
|-------|---------|-------------|
| `query` | `{ __typename }` | GraphQL query string |
| `expect_no_errors` | `false` | Fail if response contains `errors` |

---

## CLI Commands

### `rukkie login`

Log in to RukkiePulse. Required before running any other command.

```bash
rukkie login
# Password: ********
# ✅ Logged in to RukkiePulse
```

---

### `rukkie logout`

Clear the stored session.

```bash
rukkie logout
```

---

### `rukkie scan`

Scan all services: runs health checks and probes all configured endpoints concurrently.

```bash
rukkie scan
rukkie scan --env production
rukkie scan --errors-only
```

| Flag | Description |
|------|-------------|
| `--env`, `-e` | Environment to use from `rukkie.yaml` (default: `dev`) |
| `--errors-only` | Show only services with failing health or endpoints |

**Output:**

```
my-backend  [dev]

  🟢 auth-service           45ms     3/3 endpoints ok
  🔴 payment-service        —        0/2 endpoints ok  (connection refused)
  🟡 graphql-api            920ms    1/2 endpoints ok
```

---

### `rukkie status`

Alias for `rukkie scan`. Accepts the same flags.

```bash
rukkie status
rukkie status --env staging
rukkie status --errors-only
```

---

### `rukkie inspect <service>`

Deep-dive into a single service: shows health, dependency statuses, and per-endpoint results.

```bash
rukkie inspect auth-service
rukkie inspect auth-service --env production
```

**Output:**

```
Service: auth-service
URL:     http://localhost:3000
Type:    REST

Health:  🟢 ok  45ms

Dependencies:
  db:        connected
  redis:     connected

Endpoints:
  🟢  GET   /health          45ms   200
  🔴  POST  /login           120ms  500   expected 200, got 500
```

---

### `rukkie trace <service> [endpoint]`

Fetch and display the latest trace for a service from Jaeger. Requires Jaeger to be running and configured in `rukkie.yaml`.

```bash
# Latest trace for a service
rukkie trace auth-service

# Latest trace filtered to a specific endpoint/operation
rukkie trace auth-service /login

# Fetch a specific trace by ID
rukkie trace auth-service --trace-id abc123def456

# Show last 5 traces
rukkie trace auth-service /login --last 5

# Render as a flame graph instead of waterfall
rukkie trace auth-service /login --flame
```

| Flag | Description |
|------|-------------|
| `--trace-id` | Fetch a specific trace by its ID |
| `--last` | Number of recent traces to show (default: `1`) |
| `--flame` | Render as horizontal flame graph |

**Waterfall output:**

```
Service: auth-service  Endpoint: /login

Trace:  abc123def456
Total:  340ms  ❌ ERROR

  Operation                               Timeline                          Duration
  ──────────────────────────────────────────────────────────────────────────────────
  auth-service: LoginHandler              ████████████████████████████████  340ms
    auth-service: AuthService.login         ██████████████████████████      270ms
      auth-service: UserRepo.findUser           ██████████████████          170ms ❌
      │  └─ Database timeout
      auth-service: CacheService.get                        ██████          40ms

Root Cause:
  ❌ UserRepo.findUser
     Database timeout
```

---

### `rukkie watch`

Live-updating dashboard that refreshes all services on an interval.

```bash
rukkie watch
rukkie watch --interval 5s
rukkie watch --env production --interval 30s
```

| Flag | Description |
|------|-------------|
| `--interval` | Refresh interval (default: `10s`). Accepts `5s`, `1m`, etc. |
| `--env`, `-e` | Environment to use (default: `dev`) |

Press `Ctrl+C` to stop.

**Output:**

```
my-backend  [dev]   Refreshing every 10s  (Ctrl+C to stop)
Last updated: 14:23:01

  Service                   Status        Latency     Endpoints
  ────────────────────────────────────────────────────────────────────────
  auth-service              🟢 ok         45ms        3/3 pass
  payment-service           🔴 down       —           0/2 pass  ← connection refused
  graphql-api               🟡 slow       920ms       2/2 pass
  ────────────────────────────────────────────────────────────────────────

  2 ok   1 down
```

---

## Global Flags

These flags work on all commands:

| Flag | Default | Description |
|------|---------|-------------|
| `--env`, `-e` | `dev` | Environment block to use from `rukkie.yaml` |
| `--help`, `-h` | — | Show help for any command |

---

## Agent SDKs

Agents expose the `/__rukkie/health` endpoint and push OTel traces to Jaeger.

### Node.js

```bash
npm install rukkie-agent
```

```ts
import express from 'express'
import { initRukkie } from 'rukkie-agent'

const app = express()

// Call before registering routes
initRukkie({
  serviceName: 'auth-service',
  apiKey: 'rk_live_xxx',
  collectorUrl: 'http://localhost:4317',  // optional
  dependencies: {
    db: async () => checkDbConnection(),
    redis: async () => checkRedisConnection(),
  }
}, app)
```

Supports **Express** and **Fastify** (auto-detected).

---

### Python

```bash
pip install rukkie-agent
```

```python
from fastapi import FastAPI
from rukkie_agent import init_rukkie

app = FastAPI()

init_rukkie(
    service_name="auth-service",
    api_key="rk_live_xxx",
    app=app,
    collector_url="http://localhost:4317",   # optional
    dependencies={
        "db": check_db_connection,
        "redis": check_redis_connection,
    }
)
```

Supports **FastAPI** and **Flask** (auto-detected).

---

## Jaeger + OTel Collector (Docker)

Start the full observability stack locally:

```bash
# Jaeger (all-in-one)
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Jaeger UI: [http://localhost:16686](http://localhost:16686)

OTel gRPC endpoint (for agents): `http://localhost:4317`

---

## Environments

Define multiple environments in `rukkie.yaml` and switch with `--env`:

```yaml
environments:
  dev:
    services:
      - name: auth-service
        url: http://localhost:3000
        type: REST

  production:
    services:
      - name: auth-service
        url: https://auth.myapp.com
        type: REST
```

```bash
rukkie scan --env production
rukkie watch --env production
rukkie inspect auth-service --env production
```

---

## Changing the Login Password

Edit `internal/auth/credentials.go` and rebuild:

```go
const (
    hardcodedPassword = "your-new-password"
    jwtSecret         = "your-new-secret"
)
```

```bash
go build -o rukkie ./cmd/rukkie/
```
