# Phase 1 – CLI Foundation + Health Checks

## Goal

Bootstrap the Go CLI with:
- Project scaffold (Cobra CLI)
- YAML config loader
- Basic health check runner
- Terminal output

No auth, no tracing, no agents yet.

---

## Deliverables

- [ ] `rukkie scan` — hits `/__rukkie/health` on all configured services
- [ ] `rukkie status` — prints current health table
- [ ] `rukkie.yaml` — config file loaded from CWD
- [ ] Concurrent health checks via goroutines
- [ ] Clean terminal output (colored status)

---

## File Structure

```
cmd/rukkie/main.go            # Cobra root command
internal/config/config.go     # YAML loader + structs
internal/health/checker.go    # HTTP health check logic
internal/engine/engine.go     # Orchestrates concurrent checks
internal/output/printer.go    # Terminal table renderer
```

---

## Config Format (`rukkie.yaml`)

```yaml
project: my-backend

environments:
  dev:
    services:
      - name: auth-service
        url: http://localhost:3000
        type: REST

      - name: graphql-api
        url: http://localhost:4000/graphql
        type: GRAPHQL
```

---

## Config Structs (Go)

```go
// internal/config/config.go

type Config struct {
    Project      string                 `yaml:"project"`
    Environments map[string]Environment `yaml:"environments"`
}

type Environment struct {
    Services []Service `yaml:"services"`
}

type Service struct {
    Name string `yaml:"name"`
    URL  string `yaml:"url"`
    Type string `yaml:"type"` // REST | GRAPHQL
}
```

---

## Result Model

```go
// internal/engine/result.go

type Result struct {
    Service string
    Status  string        // "ok" | "degraded" | "down"
    Latency time.Duration
    Error   string
    TraceID string        // empty in Phase 1
}
```

---

## Health Checker

```
GET {service.url}/__rukkie/health
```

Expected response:

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

Rules:
- HTTP 200 + `status: "ok"` → **ok**
- HTTP 200 + `status != "ok"` → **degraded**
- HTTP non-200 or timeout → **down**
- Timeout: 5 seconds per service

---

## Concurrency Model

```go
// internal/engine/engine.go

func Run(services []config.Service) []Result {
    results := make(chan Result, len(services))
    var wg sync.WaitGroup

    for _, svc := range services {
        wg.Add(1)
        go func(s config.Service) {
            defer wg.Done()
            results <- health.Check(s)
        }(svc)
    }

    wg.Wait()
    close(results)

    var out []Result
    for r := range results {
        out = append(out, r)
    }
    return out
}
```

---

## CLI Commands

### `rukkie scan`

1. Load `rukkie.yaml` from CWD
2. Pick environment (default: `dev`)
3. Run engine concurrently
4. Print results table

### `rukkie status`

Alias for scan (same in Phase 1, will diverge later).

---

## Terminal Output Format

```
🟢 auth-service        120ms
🔴 payment-service     FAILED  (connection refused)
🟡 graphql-api         900ms   (degraded)
```

Rules:
- Green (🟢) → ok, latency < 500ms
- Yellow (🟡) → degraded OR latency >= 500ms
- Red (🔴) → down or error

---

## CLI Flag: `--env`

```bash
rukkie scan --env production
```

Selects which environment block from `rukkie.yaml` to use. Default: `dev`.

---

## Go Modules to Add

```bash
go get github.com/spf13/cobra
go get gopkg.in/yaml.v3
```

---

## Acceptance Criteria

- Running `rukkie scan` against 3 services completes in < 1 second
- Correct status shown per service (ok/degraded/down)
- Missing `rukkie.yaml` gives a clear error message
- Unknown `--env` value gives a clear error message
