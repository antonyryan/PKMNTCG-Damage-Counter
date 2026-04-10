# Backend Documentation

## Overview

The backend is a small Go application built with Gin. Its current role is intentionally simple: serve static Pokemon data through an API that supports the frontend autocomplete flow. The project structure is already prepared for future expansion, such as multiplayer synchronization, persistence, or more detailed game logic.

## File Responsibilities

- `main.go`: application entry point. It only starts the HTTP server.
- `router.go`: central route registration. This is the place to add new endpoints.
- `handlers.go`: request handlers and the reusable filtering logic used by the Pokemon search endpoint.
- `middleware.go`: cross-cutting HTTP middleware. Right now it only contains CORS handling for local development.
- `pokemon.go`: minimal domain model returned by the API.
- `pokemon_data.go`: mock Pokemon dataset used by the autocomplete endpoint.
- `main_test.go`: backend tests that validate the current API behavior.

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

Searches the in-memory Pokemon catalog by name.

Behavior:

- If `q` is empty, the endpoint returns the first 10 Pokemon from the mock list.
- If `q` is provided, the endpoint performs a case-insensitive substring match.
- Results are sorted alphabetically when a query is used.
- Responses are capped at 10 items.

Example response:

```json
[
  { "id": 4, "name": "Arceus" },
  { "id": 2, "name": "Charizard" }
]
```

## Design Notes

### Why the Pokemon list is isolated

The mock dataset lives in `pokemon_data.go` so it can be replaced later without touching routing or handler logic. This keeps the backend easier to evolve when the app moves from mock data to a real source.

### Why filtering is extracted

The `filterPokemon` function is kept separate from the HTTP handler to make the search behavior easier to test and reuse.

### Why the router is isolated

`setupRouter()` lives in `router.go` so tests can exercise the full HTTP surface without booting the real server process.

## Tests

The backend test suite currently verifies:

- empty search returns 10 items
- filtered search returns sorted results
- unmatched searches return an empty array

Run the backend checks with:

```bash
go test ./...
go build ./...
```

## Extension Guidance

Good next places for backend growth:

- move Pokemon data to JSON or database-backed storage
- add versioned API routes
- add services/repositories if business logic grows
- introduce request logging or structured configuration
