# PKMN TCG Companion

PKMN TCG Companion is a mobile-first web application designed to replace physical match accessories in in-person Pokemon TCG games. The app tracks the active Pokemon, bench slots, damage, special conditions, and once-per-game markers, while also providing quick utilities such as coin flips and die rolls.

## Project Structure

- `backend/`: Go + Gin API used to expose static Pokemon data and provide a clean backend foundation for future multiplayer features.
- `frontend/`: React + Tailwind CSS application focused on a full-screen mobile layout with offline state persistence.

## What Was Implemented

### Backend

- A modular Gin server with separated files for startup, routing, handlers, middleware, models, and mock data.
- A `GET /api/pokemon/search?q=` endpoint that filters the in-memory Pokemon list and returns at most 10 results.
- Unit tests covering empty search, filtered search, alphabetical ordering, and no-result cases.

### Frontend

- A mobile-first fixed battlefield split into two halves.
- An inverted opponent field for face-to-face play.
- Reusable Pokemon slot rendering for both active and bench positions.
- Match state management with React Context and automatic `localStorage` persistence.
- TCG-specific rules for special conditions, damage clamping, knockouts, and bench-to-active promotion.
- Utilities for coin flips, die rolls, GX markers, VSTAR markers, and full game reset.
- Unit tests for the core game-rule functions.

## Development Notes

### Backend

Run from the `backend/` directory:

```bash
go test ./...
go run main.go
```

### Frontend

Run from the `frontend/` directory:

```bash
npm install
npm test
npm run dev
```

## Additional Documentation

- `backend/README.md`: backend architecture, responsibilities, and API behavior.
- `frontend/README.md`: frontend architecture, game state design, and UI behavior.
