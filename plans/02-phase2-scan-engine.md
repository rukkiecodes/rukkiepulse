# Phase 2 – Scan Engine (REST + GraphQL Testing)

## Goal

Extend Phase 1 health checks with:
- Active endpoint testing for REST and GraphQL
- Latency profiling per endpoint
- Dependency graph checks
- `rukkie inspect <service>` command

---

## Deliverables

- [ ] `rukkie scan` — now also tests configured endpoints (not just health)
- [ ] `rukkie inspect <service>` — deep-dive into one service
- [ ] REST endpoint probe (GET/POST with expected status)
- [ ] GraphQL introspection + query probe
- [ ] Latency percentiles (p50, p95) via repeated sampling
- [ ] Dependency status shown in inspect output

---

## Config Extension

```yaml
project: my-backend

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

---

## Config Structs (Go additions)

```go
type Service struct {
    Name      string     `yaml:"name"`
    URL       string     `yaml:"url"`
    Type      string     `yaml:"type"`
    Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
    // REST
    Path         string `yaml:"path"`
    Method       string `yaml:"method"`
    Body         string `yaml:"body"`
    ExpectStatus int    `yaml:"expect_status"`

    // GraphQL
    Query          string `yaml:"query"`
    ExpectNoErrors bool   `yaml:"expect_no_errors"`
}
```

---

## Endpoint Result Model

```go
type EndpointResult struct {
    Path    string
    Method  string
    Status  string        // "pass" | "fail"
    Code    int
    Latency time.Duration
    Error   string
}
```

---

## REST Probe Logic

```
internal/probe/rest.go
```

Steps:
1. Build HTTP request (method, body, headers)
2. Record start time
3. Execute with 5s timeout
4. Record latency
5. Compare status code to `expect_status`
6. Return `EndpointResult`

---

## GraphQL Probe Logic

```
internal/probe/graphql.go
```

Steps:
1. POST to service URL with `{"query": "..."}` body
2. Parse response JSON
3. If `errors` key exists and `expect_no_errors: true` → fail
4. Return `EndpointResult`

---

## `rukkie inspect <service>` Output

```
Service: auth-service
URL:     http://localhost:3000
Type:    REST

Health:  🟢 ok (45ms)

Dependencies:
  db:    connected
  redis: connected

Endpoints:
  GET  /health     🟢  45ms    200
  POST /login      🔴  120ms   500  (Internal Server Error)
```

---

## Concurrency Extension

Each service runs its health check + all endpoint probes concurrently within a single goroutine batch.

```go
func CheckService(svc config.Service) ServiceResult {
    results := make(chan EndpointResult, len(svc.Endpoints))
    var wg sync.WaitGroup

    for _, ep := range svc.Endpoints {
        wg.Add(1)
        go func(e config.Endpoint) {
            defer wg.Done()
            results <- probe.Check(svc.URL, e)
        }(ep)
    }

    wg.Wait()
    close(results)
    // collect + return
}
```

---

## New CLI Commands

### `rukkie inspect <service>`

```bash
rukkie inspect auth-service
rukkie inspect auth-service --env production
```

Prints detailed view of one service (health + all endpoints + dependencies).

### `rukkie scan` (extended)

Now shows per-service endpoint summary:

```
🟢 auth-service        45ms    3/3 endpoints passing
🔴 payment-service     FAILED  0/2 endpoints passing
🟡 graphql-api         900ms   1/2 endpoints passing
```

---

## Go Modules to Add

No new modules needed — uses `net/http` from stdlib.

---

## Acceptance Criteria

- REST probe correctly detects 4xx/5xx failures
- GraphQL probe detects `errors` in response
- `rukkie inspect` shows per-endpoint status
- 10 endpoint probes across 5 services completes in < 2 seconds
- Partial failures (some endpoints down) correctly shown as degraded
