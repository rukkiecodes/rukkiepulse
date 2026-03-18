# Phase 6 – Tracing Visualization

## Goal

Make trace data actionable in the terminal — root cause analysis without opening a browser.

---

## Deliverables

- [ ] `rukkie trace <service> <endpoint>` — full trace waterfall in terminal
- [ ] `rukkie watch` — live-updating dashboard of all services
- [ ] Root cause highlighting (deepest failing span)
- [ ] Latency flame graph (ASCII) for a trace
- [ ] Error summary mode: `rukkie scan --errors-only`

---

## `rukkie trace` — Enhanced Waterfall

### Command

```bash
rukkie trace auth-service /login
rukkie trace auth-service /login --trace-id abc123def456
rukkie trace auth-service /login --last 5    # show last 5 traces
```

### Output

```
Trace ID: abc123def456
Service:  auth-service
Endpoint: /login
Total:    340ms  ❌ ERROR

Waterfall:
  LoginHandler                     [  0ms ──────────────────── 340ms]  ✅ 340ms
    AuthService.login              [ 10ms ──────────────── 280ms    ]  ✅ 270ms
      UserRepo.findUser            [ 30ms ──────── 200ms            ]  ❌  170ms
      │  Error: Database timeout (postgres: connection refused)
      CacheService.get             [210ms ── 250ms                  ]  ✅  40ms
    ResponseSerializer             [285ms ─ 310ms                   ]  ✅  25ms

Root Cause:
  ❌ UserRepo.findUser
     Error: Database timeout
     Hint:  Check postgres connection at localhost:5432
```

### Rendering Logic

```go
// internal/trace/renderer.go

func RenderWaterfall(trace *Trace)
func findRootCause(spans []Span) *Span
func renderBar(start, duration, total time.Duration, width int) string
```

Bar width = terminal width - label width. Normalize all spans against total trace duration.

---

## `rukkie watch` — Live Dashboard

### Command

```bash
rukkie watch
rukkie watch --env production
rukkie watch --interval 5s   # default: 10s
```

### Output (refreshes in-place)

```
RukkiePulse — my-backend [dev]          Last updated: 14:23:01

  Service             Status    Latency    Endpoints    Traces
  ──────────────────────────────────────────────────────────
  auth-service        🟢 ok      45ms       3/3 pass    0 errors
  payment-service     🔴 down    —          0/2 pass    3 errors  ← NEW
  graphql-api         🟡 slow    920ms      2/2 pass    0 errors

Press [i] to inspect  [t] to trace  [q] to quit
```

### Implementation

```go
// cmd/rukkie/watch.go

func runWatch(cfg *config.Config, env string, interval time.Duration) {
    ticker := time.NewTicker(interval)
    for range ticker.C {
        results := engine.Run(services)
        output.RenderDashboard(results)
    }
}
```

Use ANSI escape codes to clear and re-render in-place (no external TUI library needed unless we want interactivity).

Optional: use `github.com/charmbracelet/bubbletea` if keyboard interaction is desired.

---

## `rukkie scan --errors-only`

Only shows services with issues:

```bash
rukkie scan --errors-only
```

```
🔴 payment-service     FAILED    (connection refused)
🟡 graphql-api         900ms     (degraded: db slow)
```

---

## ASCII Flame Graph

Shown when running trace on a single trace ID:

```bash
rukkie trace auth-service /login --trace-id abc123 --flame
```

```
auth-service /login  340ms
█████████████████████████████████████████████████████  LoginHandler
 ████████████████████████████████████████              AuthService.login
    ██████████████████████                             UserRepo.findUser ❌
                          ██████                       CacheService.get
                                         █████         ResponseSerializer
```

Widths are proportional to duration relative to total trace time.

---

## Root Cause Detection Algorithm

```go
func findRootCause(spans []Span) *Span {
    // 1. Collect all error spans
    // 2. Find the one with no error children (leaf error)
    // 3. That is the root cause
}
```

Displayed as a separate "Root Cause" section after the waterfall.

---

## Output Package Refactor

```
internal/output/
├── printer.go      # scan table output
├── inspect.go      # service inspect output
├── dashboard.go    # watch live dashboard
└── trace.go        # waterfall + flame graph
```

---

## Go Modules to Add (optional)

```bash
# Only if we want interactive watch mode
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
```

Plain ANSI escape codes work fine for non-interactive watch.

---

## Acceptance Criteria

- `rukkie trace` waterfall renders correctly for traces with 10+ spans
- Root cause correctly identified as the deepest error span
- `rukkie watch` refreshes without screen flicker (in-place update)
- `--errors-only` filters correctly
- Flame graph widths are proportional to actual durations
- All commands degrade gracefully when Jaeger is unreachable
