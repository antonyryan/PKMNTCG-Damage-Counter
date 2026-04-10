# Frontend Documentation

## Overview

The frontend is a React + Tailwind CSS application focused on a mobile-first companion experience for physical Pokemon TCG matches. The board is designed to fill the full viewport, avoid horizontal scrolling, and remain readable from both sides of the table.

## File Responsibilities

- `src/App.tsx`: top-level composition of the battlefield and shared toolbox.
- `src/components/PokemonSlot.tsx`: reusable slot component used for active and bench positions.
- `src/context/GameContext.tsx`: central match state store, actions, and `localStorage` persistence.
- `src/context/gameRules.ts`: pure game-rule helpers used by the context and tested independently.
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

The frontend keeps one `GameState` object for the entire match.

Main concepts:

- `PlayerState`: one player board, including active slot, bench, GX marker, and VSTAR marker
- `SlotState`: one field slot, including selected Pokemon, damage, and statuses
- `GameState`: both players plus shared randomizer results for coin flips and die rolls

All state changes are persisted immediately to `localStorage` under a single key so reloading the browser restores the current match.

## Game Rules Implemented

### Pokemon management

- empty slots can be filled through autocomplete
- knockout clears the slot completely
- promoting a bench Pokemon swaps it with the active Pokemon if one exists
- when the former active Pokemon moves to the bench, its special conditions are cleared

### Special conditions

The active slot supports the official TCG grouping used in this app:

- `sleep`, `confusion`, and `paralysis` are mutually exclusive
- `poison` and `burn` can coexist with each other and with one exclusive status

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

It is intentionally reusable so the active slot and all bench slots share the same rendering model.

## Tests

The frontend test suite currently covers the pure game rules in `src/context/gameRules.test.ts`.

Covered behavior:

- exclusive status replacement
- poison and burn toggling
- damage floor at zero
- bench promotion with swap and status cleanup
- empty bench promotion as a no-op

Run the frontend checks with:

```bash
npm install
npm test
npm run build
```

## Extension Guidance

Good next places for frontend growth:

- add component integration tests for `PokemonSlot`
- replace the mock autocomplete source with official card data
- add animations for status changes and knockouts
- add a multiplayer transport layer when backend synchronization exists
