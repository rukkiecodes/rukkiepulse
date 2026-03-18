# Phase 4 – OpenTelemetry Integration

## Goal

Wire the CLI to an OpenTelemetry-compatible tracing backend (Jaeger) so that:
- The CLI can fetch traces for a given service/endpoint
- Trace IDs surfaced in health/scan results link to full trace trees
- `rukkie trace <service> <endpoint>` shows the trace waterfall in terminal

---

## Deliverables

- [ ] OTel trace fetcher (queries Jaeger HTTP API)
- [ ] `rukkie trace <service> <endpoint>` command
- [ ] Trace ID attached to `Result` struct (from scan/inspect)
- [ ] Trace waterfall rendered in terminal
- [ ] Jaeger config block in `rukkie.yaml`

---

## Architecture

```
CLI ──► Jaeger Query API (HTTP)
          GET /api/traces?service=auth-service&...
          GET /api/traces/{traceID}
```

The CLI does NOT produce traces — it only consumes them from Jaeger.
Services + agents produce traces (Phase 5).

---

## Config Extension (`rukkie.yaml`)

```yaml
observability:
  jaeger:
    url: http://localhost:16686
```

---

## Config Struct Extension

```go
type Config struct {
    Project       string                 `yaml:"project"`
    Environments  map[string]Environment `yaml:"environments"`
    Observability Observability          `yaml:"observability"`
}

type Observability struct {
    Jaeger JaegerConfig `yaml:"jaeger"`
}

type JaegerConfig struct {
    URL string `yaml:"url"`
}
```

---

## Trace Package Structure

```
internal/trace/
├── client.go     # HTTP client for Jaeger Query API
├── model.go      # Trace + Span structs
└── renderer.go   # Terminal waterfall printer
```

---

## Trace Model

```go
// internal/trace/model.go

type Trace struct {
    TraceID string
    Spans   []Span
}

type Span struct {
    SpanID        string
    ParentSpanID  string
    OperationName string
    Service       string
    StartTime     time.Time
    Duration      time.Duration
    Status        string   // "ok" | "error"
    Error         string
}
```

---

## Jaeger Client

```go
// internal/trace/client.go

type Client struct {
    BaseURL string
    Token   string
}

// Fetch most recent traces for a service+endpoint
func (c *Client) FetchTraces(service, endpoint string, limit int) ([]Trace, error)

// Fetch one trace by ID
func (c *Client) FetchTrace(traceID string) (*Trace, error)
```

Jaeger Query API endpoints used:

```
GET /api/traces?service={service}&operation={endpoint}&limit={n}
GET /api/traces/{traceID}
```

---

## Result Model Extension

```go
// internal/engine/result.go

type Result struct {
    Service  string
    Status   string
    Latency  time.Duration
    Error    string
    TraceID  string   // now populated from latest Jaeger trace
}
```

After a scan, if Jaeger is configured, the engine fetches the latest trace ID for each service and attaches it.

---

## `rukkie trace <service> <endpoint>` Command

```bash
rukkie trace auth-service /login
```

Steps:
1. Load config (get Jaeger URL)
2. Authenticate (load JWT)
3. Call `client.FetchTraces("auth-service", "/login", 1)`
4. Render waterfall of most recent trace

---

## Terminal Waterfall Renderer

```go
// internal/trace/renderer.go

func Render(trace *Trace)
```

Output:

```
Trace: abc123def456
Service: auth-service  Endpoint: /login  Duration: 340ms

  LoginHandler                           0ms    120ms  ✅
    → AuthService.login                  5ms     80ms  ✅
      → UserRepo.findUser               10ms     60ms  ❌
          Error: Database timeout
```

Rules:
- Indent spans by parent depth
- Show relative start offset + duration
- ✅ = no error, ❌ = has error
- Highlight error spans in red

---

## Scan Integration

When Jaeger is configured, `rukkie scan` appends trace IDs to results:

```
🔴 payment-service  FAILED  trace: 9f3c2a1b...
```

User can then run:

```bash
rukkie trace payment-service /charge
```

---

## Go Modules to Add

```bash
go get go.opentelemetry.io/otel
```

Note: The CLI itself does not instrument — no tracer setup needed in the CLI.
The `go.opentelemetry.io/otel` dependency is only for shared type definitions if needed.
Jaeger is queried via plain HTTP calls.

---

## Local Dev Setup (Jaeger)

```bash
docker run -d --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  jaegertracing/all-in-one:latest
```

Jaeger UI: http://localhost:16686

---

## Acceptance Criteria

- `rukkie trace auth-service /login` renders a trace waterfall
- Error spans highlighted and cause displayed
- Trace ID shown in scan output when Jaeger is configured
- No crash if Jaeger is unreachable (graceful fallback, warning printed)
- Waterfall correctly indents nested spans by depth
