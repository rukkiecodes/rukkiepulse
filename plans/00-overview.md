# RukkiePulse – Implementation Overview

## What We're Building

A CLI-first, multi-language observability and diagnostics platform.

```
Rukkie CLI (Go)  ──►  services (via HTTP health checks)
                  ◄──  agents push telemetry (OTel/Jaeger)
```

---

## Repo Structure

```
rukkiePulse/
├── cmd/                    # CLI entry points
│   └── rukkie/
│       └── main.go
├── internal/
│   ├── auth/               # JWT login, token storage
│   ├── config/             # YAML config loader
│   ├── engine/             # Scan, check, aggregate
│   ├── health/             # Health check runner
│   ├── trace/              # OTel trace fetcher
│   └── output/             # Terminal renderer
├── agents/
│   ├── node/               # Node.js rukkie-agent SDK
│   └── python/             # Python rukkie-agent SDK
├── plans/                  # This folder
└── rukkie.yaml             # Example config
```

---

## Tech Stack

| Layer         | Technology                        |
|---------------|-----------------------------------|
| CLI           | Go + Cobra                        |
| Config        | YAML (`gopkg.in/yaml.v3`)         |
| Auth          | JWT (`golang-jwt/jwt`)            |
| Concurrency   | Go goroutines + channels          |
| Observability | OpenTelemetry Go SDK              |
| Tracing UI    | Jaeger                            |
| Node Agent    | TypeScript + OTel Node SDK        |
| Python Agent  | Python + OTel Python SDK          |
| Terminal UI   | `github.com/charmbracelet/lipgloss` or plain ANSI |

---

## Development Phases

| Phase | Focus                              | Plan File                        |
|-------|------------------------------------|----------------------------------|
| 1     | CLI foundation + health checks     | `01-phase1-cli-foundation.md`    |
| 2     | Scan engine (REST + GraphQL)       | `02-phase2-scan-engine.md`       |
| 3     | Auth (JWT + API keys)              | `03-phase3-auth.md`              |
| 4     | OpenTelemetry integration          | `04-phase4-opentelemetry.md`     |
| 5     | Node.js & Python agent SDKs        | `05-phase5-agents.md`            |
| 6     | Tracing visualization              | `06-phase6-tracing-viz.md`       |

---

## Key Constraints

- Support 50+ services
- Full scan must complete in < 3 seconds (concurrency is critical)
- Language-agnostic agents
- Minimal setup for developers

---

## Global Principles

1. Build in phases — do not add Phase 3+ code during Phase 1
2. Keep code modular — each `internal/` package owns one concern
3. Strong Go typing — use structs, not `map[string]interface{}`
4. Reusable agent SDKs — both Node and Python agents share the same contract
