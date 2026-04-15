# Frontend Documentation

## Overview

The frontend is a React + Tailwind CSS application focused on a mobile-first companion experience for physical Pokemon TCG matches. The board is designed to fill the full viewport, avoid horizontal scrolling, and remain readable from both sides of the table.

## File Responsibilities

- `src/App.tsx`: top-level composition of the battlefield and shared toolbox.
- `src/components/PokemonSlot.tsx`: reusable slot component used for active and bench positions.
- `src/context/GameContext.tsx`: central match store backed by the backend session API.
- `src/api/client.ts`: frontend API layer for catalog lookup and session actions.
- `src/types.ts`: shared TypeScript model definitions for the whole app.
- `src/index.css`: global styles, Tailwind setup, and full-screen layout defaults.
- `src/main.tsx`: React application bootstrap.

## Layout Model

The battlefield is split into two vertical halves:

- upper half: opponent field, rotated 180 degrees
- lower half: local player field, normal orientation

Each side contains:

- 1 active slot
- 5 bench slots

A floating toolbox sits in the middle of the screen so both players can access shared match actions.

## State Model

The frontend keeps the current backend session snapshot in memory and only persists the `sessionId` locally.

Main concepts:

- `PlayerState`: one player board, including active slot, bench, GX marker, and VSTAR marker
- `SlotState`: one field slot, including selected Pokemon, damage, and statuses
- `GameState`: both players plus shared randomizer results for coin flips and die rolls
- `GameHistoryEntry`: persisted backend action log exposed to the UI when needed

All gameplay changes are sent to the backend as validated actions. Reloading the browser restores the authoritative session snapshot by reusing the stored `sessionId`.

## Backend-Driven Interactions

### Pokemon management

- empty slots can be filled through backend autocomplete
- evolve and de-evolve choices come from the backend per selected Pokemon
- knockout, bench promotion, and all other board mutations are sent as backend actions

### Persistence

- the browser stores only the backend `sessionId`
- the backend persists the full match snapshot and action history

### Match utilities

The shared toolbox provides:

- coin flip with a short animation
- 1 to 6 die roll
- GX toggle for each side
- VSTAR toggle for each side
- full match reset

## Component Behavior

### `PokemonSlot`

This component renders two modes:

- empty mode: shows an add button and the autocomplete input flow
- occupied mode: shows Pokemon name, damage controls, knockout, and optional active-only or bench-only actions

It is intentionally reusable so the active slot and all bench slots share the same rendering model, while delegating add and evolve lookups to the backend.

## Tests

The frontend test suite currently covers the API client helpers in `src/api/client.test.ts`.

Run the frontend checks with:

```bash
npm install
npm run test -- --run
npm run build
```

## Extension Guidance

Good next places for frontend growth:

- add component integration tests for `PokemonSlot`
- add UI for browsing backend session history
- add animations for status changes and knockouts
- add analytics views backed by backend history data
