# PKMN TCG Companion

PKMN TCG Companion is a mobile-first web application designed to replace physical match accessories in in-person Pokemon TCG games. The app tracks the active Pokemon, bench slots, damage, special conditions, and once-per-game markers, while also providing quick utilities such as coin flips and die rolls.

## Project Structure

- `backend/`: Go + Gin API that owns the Pokemon catalog, evolution validation, match rules, session persistence, and action history.
- `frontend/`: React + Tailwind CSS application focused on a full-screen mobile layout backed by the backend session API.
- `pokemon_data.json`: authoritative Pokemon catalog consumed by the backend.

## What Was Implemented

### Backend

- A modular Gin server with separated files for startup, routing, handlers, catalog loading, game rules, and session storage.
- Catalog search and per-Pokemon evolution-option endpoints backed by the root JSON dataset.
- Session endpoints that create or restore a match, apply validated actions, and expose persisted history.
- Unit tests covering catalog search, full-chain evolution options, and session action validation.

### Frontend

- A mobile-first fixed battlefield split into two halves.
- An inverted opponent field for face-to-face play.
- Reusable Pokemon slot rendering for both active and bench positions.
- Match state management with React Context backed by the backend session API.
- Local persistence limited to the backend `sessionId`, so reloads restore the authoritative server-side snapshot.
- Utilities for coin flips, die rolls, GX markers, VSTAR markers, and full game reset.
- Backend-driven add and evolve flows for active and bench slots.
- Unit tests for the API client helpers.

## Development Notes

### Backend

Run from the `backend/` directory:

```bash
go test ./...
go build ./...
go run main.go
```

### Frontend

Run from the `frontend/` directory:

```bash
npm install
npm run test -- --run
npm run dev
npm run build
```

## Additional Documentation

- `backend/README.md`: backend architecture, responsibilities, and API behavior.
- `frontend/README.md`: frontend architecture, game state design, and UI behavior.
