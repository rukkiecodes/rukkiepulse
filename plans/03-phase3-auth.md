# Phase 3 – Authentication (JWT + API Keys)

## Goal

Add a secure identity layer:
- `rukkie login` stores a JWT for the CLI user
- Agent SDKs authenticate via per-service API keys
- All CLI→backend requests carry a Bearer JWT
- All agent→backend requests carry `x-rukkie-api-key`

---

## Deliverables

- [ ] `rukkie login` — prompts for credentials, exchanges for JWT, stores locally
- [ ] `rukkie logout` — removes stored token
- [ ] JWT loaded automatically on every CLI command
- [ ] API key format defined and validated
- [ ] CLI enforces auth before running scans (no token → error + hint)
- [ ] Token refresh logic (if JWT expires)

---

## Token Storage

```
~/.rukkie/config.yaml
```

```yaml
token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
expires_at: 2026-04-01T00:00:00Z
```

---

## Auth Package Structure

```
internal/auth/
├── login.go       # Credential exchange → JWT
├── store.go       # Read/write ~/.rukkie/config.yaml
├── token.go       # Load, validate, refresh token
└── apikey.go      # API key format + validation helpers
```

---

## Login Flow

Credentials are hard-coded in the CLI (single-user tool). No auth server or database needed.

```
rukkie login
> Password: ********

✅ Logged in
```

### `internal/auth/login.go`

```go
// Hard-coded credentials — single user only
const hardcodedPassword = "your-password-here"

func Login(password string) error {
    if password != hardcodedPassword {
        return errors.New("invalid password")
    }
    // generate a local JWT signed with a static secret
    // store to ~/.rukkie/config.yaml
    return nil
}
```

JWT is signed locally with a static secret (no server). It just proves the user ran `rukkie login` with the right password.

---

## Token Loading (on every command)

```go
// internal/auth/token.go

func LoadToken() (string, error) {
    cfg, err := store.Read()
    if err != nil || cfg.Token == "" {
        return "", errors.New("not authenticated — run `rukkie login`")
    }
    if time.Now().After(cfg.ExpiresAt) {
        return "", errors.New("session expired — run `rukkie login`")
    }
    return cfg.Token, nil
}
```

Cobra middleware calls `auth.LoadToken()` before any protected command runs.

---

## Request Headers

All HTTP requests made by the CLI must include:

```http
Authorization: Bearer <JWT>
```

Add a shared HTTP client wrapper:

```go
// internal/httpclient/client.go

func New(token string) *http.Client

func AuthRequest(method, url, token string, body io.Reader) (*http.Request, error)
```

---

## API Key System

### Format

```
rk_live_<32-char-hex>
rk_test_<32-char-hex>
```

- `rk_live_` — production
- `rk_test_` — non-production

### Validation Helper

```go
// internal/auth/apikey.go

func ValidateAPIKey(key string) bool {
    // must match: rk_(live|test)_[a-f0-9]{32}
}
```

### Agent Request Header

Agents send:

```http
x-rukkie-api-key: rk_live_abc123...
```

This is handled in Phase 5 (agent SDKs), but the validation logic lives here.

---

## Config YAML Extension

Allow API key to be stored per-service in `rukkie.yaml` for local dev:

```yaml
environments:
  dev:
    services:
      - name: auth-service
        url: http://localhost:3000
        type: REST
        api_key: rk_test_abc123  # optional — for agent identity verification
```

---

## Config Struct Extension

```go
type Service struct {
    Name      string     `yaml:"name"`
    URL       string     `yaml:"url"`
    Type      string     `yaml:"type"`
    APIKey    string     `yaml:"api_key"`
    Endpoints []Endpoint `yaml:"endpoints"`
}
```

---

## CLI Commands

### `rukkie login`

```bash
rukkie login
```

Prompts for email + password, fetches JWT, stores to `~/.rukkie/config.yaml`.

### `rukkie logout`

```bash
rukkie logout
```

Deletes token from `~/.rukkie/config.yaml`.

---

## Protected Commands

After Phase 3, these commands require a valid JWT:

- `rukkie scan`
- `rukkie status`
- `rukkie inspect`
- `rukkie trace`
- `rukkie watch`

Unprotected (public):

- `rukkie login`

---

## Go Modules to Add

```bash
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/term   # for masked password input
```

---

## Acceptance Criteria

- `rukkie login` stores a valid token
- Running `rukkie scan` without a token prints a clear error
- Expired token triggers helpful re-login prompt
- API key validation correctly rejects malformed keys
- All outbound HTTP requests include `Authorization: Bearer ...`
