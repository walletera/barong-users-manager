# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build
go build ./...

# Run all tests
make test_all

# Run a single scenario
go test -count=1 -v ./internal/tests/... -run TestUserCreatedEventProcessing
```

Tests require Docker to spin up RabbitMQ and MockServer containers via testcontainers-go.

## Architecture

This is a **Barong user management microservice**. It consumes JWS events from RabbitMQ and calls the Barong Admin API to update user state.

**Event flow:**
1. Barong publishes a JWS message to exchange `barong.events.model`, routing key `user.created`
2. The deserializer base64-decodes the `payload` field (`base64.RawURLEncoding`) and checks `event.name`
3. The handler extracts `event.record.uid` and calls `PUT /api/v1/auth/admin/users` with `state=active`

### Layers

- **`cmd/main.go`** — Entry point; reads env vars, wires `app.NewApp`, handles SIGINT/SIGTERM
- **`internal/app/`** — Bootstrap: logs in to Barong at startup, creates admin client, wires RabbitMQ consumer and message processor
- **`internal/domain/events/barong/`** — Event types, deserializer, and handler
  - `user_created_event.go` — `UserCreated` struct + `Handler` interface
  - `deserializer.go` — Unwraps JWS envelope, decodes payload, returns `UserCreated` or nil
  - `events_handler.go` — Calls `admin.Client.UpdateUser`; treats `admin.user.state_no_change` (422) as idempotent success
- **`internal/tests/`** — BDD integration tests (godog + MockServer)

### External Dependencies

| Dependency | Purpose |
|---|---|
| RabbitMQ | Consumes events from exchange `barong.events.model`, routing key `user.created`, queue `barong-users-manager` |
| Barong Admin API | `PUT /api/v1/auth/admin/users` — sets user `state=active` |
| Barong User API | `POST /api/v1/auth/identity/sessions` — login at startup to obtain session cookies |

### Required Environment Variables

```
RABBITMQ_HOST, RABBITMQ_PORT, RABBITMQ_USER, RABBITMQ_PASSWORD
BARONG_URL            # Base URL of the Barong instance (e.g. http://barong:9090)
BARONG_ADMIN_EMAIL    # Email of the Barong admin account
BARONG_ADMIN_PASSWORD # Password of the Barong admin account
```

## Testing

Tests are BDD-style using [godog](https://github.com/cucumber/godog) with Gherkin feature files in `internal/tests/features/`.

- `TestMain` starts real Docker containers: RabbitMQ (`rabbitmq-bum`) and MockServer (`mockserver-bum`) on port 2090
- Test state is passed via `context.WithValue`; log watching via `slogwatcher` asserts service behavior
- MockServer mocks both the Barong login endpoint and the update-user endpoint

**Login mock ordering** — The Barong login mock must be registered in the Gherkin `Background` *before* the "a running barong-users-manager" step, because `App.Run()` calls the login during startup.

**Idempotency** — Barong returns `422 {"errors":["admin.user.state_no_change"]}` when `state=active` is set on an already-active user. The handler detects this via `strings.Contains` and treats it as a non-retryable success (acks the message normally).

**Session expiry** — The service logs in once at startup and reuses session cookies for its lifetime. Re-login on 401 is a known follow-up task.
