# Backend Documentation

## Overview

The backend is a Go + Gin application that acts as the source of truth for the companion app. It owns the Pokemon catalog, evolution validation, match-state mutations, persisted sessions, and action history exposed to the frontend.

## File Responsibilities

- `main.go`: application entry point. It only starts the HTTP server.
- `router.go`: central route registration. This is the place to add new endpoints.
- `handlers.go`: request handlers for catalog lookup, session restore, and action application.
- `middleware.go`: cross-cutting HTTP middleware. Right now it only contains CORS handling for local development.
- `pokemon.go`: transport types returned by catalog and evolution endpoints.
- `catalog.go`: JSON catalog loader plus search and evolution graph queries.
- `domain.go`: backend-owned match state and validated rule application.
- `session_store.go`: persisted session snapshots and history storage.
- `main_test.go`: backend tests that validate the API behavior.

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

### `GET /api/sessions/:id`

Returns the authoritative current snapshot for one match session.

### `GET /api/sessions/:id/history`

Returns the persisted action log for one match session.

### `POST /api/sessions/:id/actions`

Applies a validated action to the current session and persists both state and history.

Example response:

```json
[
  { "id": 4, "name": "Arceus" },
  { "id": 2, "name": "Charizard" }
]
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

Run the backend checks with:

```bash
go test ./...
go build ./...
```

## Extension Guidance

Good next places for backend growth:

- add versioned API routes
- add analytics endpoints over persisted session history
- add services/repositories if business logic grows further
- introduce request logging or structured configuration
