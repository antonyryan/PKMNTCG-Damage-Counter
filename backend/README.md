# Backend Documentation

## Overview

The backend is a Go + Gin application that acts as the source of truth for the companion app. It owns the Pokemon catalog, evolution validation, match-state mutations, persisted sessions, and action history exposed to the frontend.

## Persistence Architecture

The following contracts are binding for all future development. Violating them constitutes an architectural regression.

| Data | Storage | Rationale |
|------|---------|-----------|
| **Analytics** (pokemon usage, damage, knockouts) | SQLite (`analytics.db`) via worker/channel queue | Durable metrics; worker pattern ensures serialised access without SQLITE_BUSY errors |
| **Sessions** (match state + action history) | RAM cache + JSON files on disk — no SQLite | Low-latency reads; JSON snapshots provide crash recovery |
| **Pokemon catalog** | In-memory only — loaded from bundled JSON at startup | Immutable reference data; no persistence needed |
| **Per-request game state** | Transient in-memory within each HTTP handler | Discarded after response |
| **Frontend state** | React context in memory; `sessionId` in `localStorage` | Restored from backend on page load |

### SQLite access rule

Every new SQLite operation — read or write — must go through the worker/channel layer in `internal/analytics`. Direct `db.Exec` / `db.Query` calls outside that worker goroutine are forbidden. This prevents concurrency bugs and keeps DB lifecycle management in one place.

## Module Architecture (Feature-First)

The backend now follows a feature-first internal package layout:

- `internal/httpapi`: transport layer (Gin handlers, route registration, middleware)
- `internal/session`: game domain + action rules + session persistence (RAM + JSON)
- `internal/catalog`: catalog ingestion and search/evolution rules
- `internal/analytics`: SQLite worker/channel, schema and query logic
- `internal/app`: composition root wiring dependencies and lifecycle

Dependency rule (must not be violated):

1. HTTP (`internal/httpapi`) depends on application services.
2. Application services depend on domain contracts/types.
3. Domain depends only on its own rules/contracts.
4. Infrastructure adapters (SQLite/filesystem) stay encapsulated inside feature packages.

## File Responsibilities

- `main.go`: process entrypoint, signal handling, graceful server shutdown.
- `internal/app/app.go`: composition root and dependency injection.
- `internal/httpapi/router.go`: route contracts and registration.
- `internal/httpapi/handlers.go`: feature handlers without global state.
- `internal/httpapi/middleware.go`: shared HTTP middleware.
- `internal/session/*`: session domain types, action rules, and JSON-backed store.
- `internal/catalog/*`: Pokemon catalog models + loader + search/evolution logic.
- `internal/analytics/*`: SQLite worker/channel store, schema, and analytics queries.
- `main_test.go`: backend tests that validate the API behavior, including concurrency tests for the analytics worker.

## Current API

### `GET /health`

Returns a basic JSON payload to confirm the API is running.

Example response:

```json
{
  "status": "ok"
}
```

### `GET /api/pokemon/search?q={query}`

Searches the backend-owned Pokemon catalog by name.

Behavior:

- If `q` is empty, the endpoint returns Pokemon in ascending National Dex ID order.
- If `q` is provided, the endpoint performs a case-insensitive substring match.
- Responses are capped by `limit` and default to 20 items.

### `GET /api/pokemon/:id/evolution-options?q={query}`

Returns only valid evolution and de-evolution targets for the given Pokemon.

Behavior:

- Traverses the full evolution chain, so stage jumps in either direction are supported.
- Labels each result as `Evolve` or `De-evolve`.
- Filters by optional case-insensitive substring query.

### `POST /api/sessions`

Creates a new session or restores an existing one when `sessionId` is provided.

```bash
# New session
curl -s -X POST http://localhost:8080/api/sessions \
  -H 'Content-Type: application/json' -d '{}'

# Restore
curl -s -X POST http://localhost:8080/api/sessions \
  -H 'Content-Type: application/json' \
  -d '{"sessionId": "<id>"}'
```

### `GET /api/sessions/:id`

Returns the authoritative current snapshot for one match session.

```bash
curl http://localhost:8080/api/sessions/<id>
```

### `GET /api/sessions/:id/history`

Returns the persisted action log for one match session.

```bash
curl http://localhost:8080/api/sessions/<id>/history
```

### `POST /api/sessions/:id/actions`

Applies a validated action to the current session and persists both state and history.

Supported `type` values and their required fields:

| type             | required fields                                                    |
| ---------------- | ------------------------------------------------------------------ |
| `set-pokemon`    | `side`, `zone`, `pokemonId` (+ `benchIndex` if zone=bench)         |
| `evolve-pokemon` | `side`, `zone`, `pokemonId` (+ `benchIndex` if zone=bench)         |
| `knockout`       | `side`, `zone` (+ `benchIndex` if zone=bench)                      |
| `adjust-damage`  | `side`, `zone`, `amount` (+ `benchIndex` if zone=bench)            |
| `toggle-status`  | `side`, `status` (`sleep`/`confusion`/`paralysis`/`poison`/`burn`) |
| `promote-bench`  | `side`, `benchIndex`                                               |
| `toggle-gx`      | `side`                                                             |
| `toggle-vstar`   | `side`                                                             |
| `flip-coin`      | _(none)_                                                           |
| `roll-die`       | _(none)_                                                           |
| `reset`          | _(none)_                                                           |

```bash
curl -s -X POST http://localhost:8080/api/sessions/<id>/actions \
  -H 'Content-Type: application/json' \
  -d '{"type":"set-pokemon","side":"me","zone":"active","pokemonId":25}'
```

### `GET /api/analytics/pokemon-usage?limit={n}`

Returns the most-used Pokémon sorted by usage count descending.

```bash
curl "http://localhost:8080/api/analytics/pokemon-usage?limit=10"
```

Example response:

```json
{
  "pokemon": [
    { "pokemonId": 25, "name": "Pikachu", "useCount": 14 },
    { "pokemonId": 6, "name": "Charizard", "useCount": 9 }
  ]
}
```

### `GET /api/analytics/damage`

Returns total damage dealt and total intentional healing across all sessions.
Damage removals within 2 seconds of the previous damage event on the same slot are excluded (treated as corrections).

```bash
curl http://localhost:8080/api/analytics/damage
```

Example response:

```json
{ "totalDealt": 2480, "totalHealed": 120 }
```

### `GET /api/analytics/knockouts`

Returns the total knockout count across all sessions.

```bash
curl http://localhost:8080/api/analytics/knockouts
```

Example response:

```json
{ "totalKnockouts": 37 }
```

## Design Notes

### Catalog source of truth

The backend loads the catalog from the repository-root `pokemon_data.json` file. This keeps frontend bundles smaller and ensures search and evolution validation use the same dataset.

### Session authority

The frontend sends actions, but the backend validates and applies them. This keeps evolution logic, damage/status rules, and restore behavior centralized.

### Router isolation

`setupRouter()` lives in `router.go` so tests can exercise the full HTTP surface without booting the real server process.

## Tests

The backend test suite currently verifies:

- empty search returns catalog data in ID order
- filtered search returns matching Pokemon only
- evolution options cover full-chain evolve and de-evolve transitions
- session actions validate evolution targets and persist history

### Official Commands

Use these commands from [backend](backend):

```bash
go test .
go test ./...
go test -v ./...
go build ./...
```

Command intent:

- `go test .`: quick smoke test for the current package.
- `go test ./...`: full backend coverage across all packages.
- `go test -v ./...`: full suite with RUN/PASS/FAIL visibility (recommended in local debugging).

Note: by Go toolchain design, plain `go test` runs only the current package. For total backend execution, use `go test ./...` (or `gotestsum ... ./...`).

### Pretty Output with gotestsum

To improve terminal readability, use gotestsum as the presentation layer while keeping `go test` semantics:

```bash
gotestsum --format testname -- -v ./...
```

If gotestsum is not installed, fallback remains:

```bash
go test -v ./...
```

From repository root, npm scripts are available:

```bash
npm run backend:test:smoke
npm run backend:test:all
npm run backend:test:verbose
npm run backend:test:pretty
```

### Naming and Log Conventions

- Test names use `Feature_Cenario_ResultadoEsperado`.
- Use `t.Run` for sub-scenarios so verbose mode stays easy to scan.
- Use `t.Logf` only when context helps diagnose failures.

## Extension Guidance

Good next places for backend growth:

- add versioned API routes
- add per-session analytics breakdown
- add services/repositories if business logic grows further
- introduce request logging or structured configuration
