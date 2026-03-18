# Phase 5 – Agent SDKs (Node.js + Python)

## Goal

Build lightweight, zero-config SDK packages that developers drop into their services to:
- Auto-instrument HTTP requests
- Expose `/__rukkie/health` endpoint
- Push metrics, errors, and traces to the OTel collector

---

## Deliverables

- [ ] `agents/node/` — TypeScript package (`rukkie-agent`)
- [ ] `agents/python/` — Python package (`rukkie-agent`)
- [ ] Both auto-detect framework (Express, Fastify / FastAPI, Flask)
- [ ] Both expose `/__rukkie/health`
- [ ] Both instrument requests with OTel spans
- [ ] Both push metrics to OTel Collector via OTLP
- [ ] Shared data contract (metrics + error payload format)

---

## Shared Contract

### Health Response

Both agents expose:

```
GET /__rukkie/health
```

```json
{
  "status": "ok",
  "service": "auth-service",
  "dependencies": {
    "db": "connected",
    "redis": "connected"
  }
}
```

### Metrics Payload (OTLP)

Sent via OTel SDK — not manually:

```json
{
  "service": "auth-service",
  "endpoint": "/login",
  "duration": 120,
  "status": 500
}
```

### Error Payload

Captured as OTel span events:

```json
{
  "service": "auth-service",
  "error": "Database timeout",
  "traceId": "abc123"
}
```

---

## Node.js Agent

### Location

```
agents/node/
├── src/
│   ├── index.ts          # Public API: initRukkie()
│   ├── otel.ts           # OTel SDK setup (tracer, meter)
│   ├── middleware.ts      # Express/Fastify middleware
│   ├── health.ts          # /__rukkie/health route
│   └── detect.ts          # Framework auto-detection
├── package.json
└── tsconfig.json
```

### Public API

```ts
import { initRukkie } from "rukkie-agent";

initRukkie({
  serviceName: "auth-service",
  apiKey: "rk_live_xxx",
  collectorUrl: "http://localhost:4317",  // optional, default: localhost
  dependencies: {                          // optional
    db: () => checkDbConnection(),
    redis: () => checkRedisConnection()
  }
});
```

### Framework Detection (`detect.ts`)

```ts
function detectFramework(app: any): "express" | "fastify" | "unknown"
```

- Checks for `app._router` → Express
- Checks for `app.routing` → Fastify
- Falls back to manual registration if unknown

### OTel Setup (`otel.ts`)

```ts
function setupOtel(config: RukkieConfig): void {
    // 1. NodeSDK with OTLP gRPC exporter
    // 2. Auto-instrument HTTP, express/fastify
    // 3. Register metrics (request count, duration histogram)
}
```

### Middleware (`middleware.ts`)

Wraps every request:

1. Start OTel span with `http.method`, `http.route`, `http.status_code`
2. Record request duration in histogram
3. On error: record exception on span, set span status ERROR

### Health Endpoint (`health.ts`)

```ts
async function buildHealthResponse(
  serviceName: string,
  dependencies: DependencyChecks
): Promise<HealthResponse>
```

Runs all dependency check functions, returns `{ status: "ok" | "degraded", ... }`.

### Packages

```json
{
  "dependencies": {
    "@opentelemetry/sdk-node": "^0.52.0",
    "@opentelemetry/exporter-trace-otlp-grpc": "^0.52.0",
    "@opentelemetry/auto-instrumentations-node": "^0.48.0"
  }
}
```

---

## Python Agent

### Location

```
agents/python/
├── rukkie_agent/
│   ├── __init__.py        # Public API: init_rukkie()
│   ├── otel.py            # OTel SDK setup
│   ├── middleware.py       # ASGI/WSGI middleware
│   ├── health.py           # /__rukkie/health handler
│   └── detect.py           # Framework auto-detection
├── pyproject.toml
└── README.md
```

### Public API

```python
from rukkie_agent import init_rukkie

init_rukkie(
    app=app,
    service_name="auth-service",
    api_key="rk_live_xxx",
    collector_url="http://localhost:4317",   # optional
    dependencies={                            # optional
        "db": check_db_connection,
        "redis": check_redis_connection
    }
)
```

### Framework Detection (`detect.py`)

```python
def detect_framework(app) -> str:
    # FastAPI: isinstance(app, FastAPI)
    # Flask:   isinstance(app, Flask)
    # Starlette: isinstance(app, Starlette)
    # fallback: "unknown"
```

### OTel Setup (`otel.py`)

```python
def setup_otel(config: RukkieConfig) -> None:
    # 1. TracerProvider with OTLP gRPC exporter
    # 2. MeterProvider with OTLP exporter
    # 3. Auto-instrument HTTP calls (requests, httpx)
    # 4. Framework-specific instrumentation
```

### Middleware (`middleware.py`)

ASGI middleware (works with FastAPI/Starlette):

```python
class RukkieMiddleware:
    async def __call__(self, scope, receive, send):
        # start span
        # call next
        # record duration + status code
        # record error if exception
```

Flask uses a WSGI equivalent or `before_request`/`after_request` hooks.

### Health Endpoint (`health.py`)

```python
async def health_handler(service_name: str, dependencies: dict) -> dict:
    # run each dependency check
    # return {"status": "ok"|"degraded", "service": ..., "dependencies": {...}}
```

Registered as a route at `/__rukkie/health` automatically.

### Packages

```toml
[project.dependencies]
opentelemetry-sdk = ">=1.24"
opentelemetry-exporter-otlp-proto-grpc = ">=1.24"
opentelemetry-instrumentation-fastapi = ">=0.45"
opentelemetry-instrumentation-flask = ">=0.45"
opentelemetry-instrumentation-requests = ">=0.45"
```

---

## OTel Collector Config (local dev)

```yaml
# otel-collector.yaml
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  jaeger:
    endpoint: jaeger:14250
    tls:
      insecure: true

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [jaeger]
```

```bash
docker run -p 4317:4317 \
  -v $(pwd)/otel-collector.yaml:/etc/otel-collector.yaml \
  otel/opentelemetry-collector:latest \
  --config /etc/otel-collector.yaml
```

---

## Acceptance Criteria

- `initRukkie()` / `init_rukkie()` works with zero-config (just service name + API key)
- `/__rukkie/health` returns correct status within 200ms
- Every HTTP request creates an OTel span visible in Jaeger
- Exception on a route creates an error span with the message attached
- Dependency check failures correctly set status to "degraded"
- Works with Express, Fastify, FastAPI, Flask without code changes beyond the init call
