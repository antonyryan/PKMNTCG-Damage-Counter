// NonStackingStatus covers the mutually exclusive active-only conditions in the TCG.
export type NonStackingStatus = "sleep" | "confusion" | "paralysis";
// StackingStatus covers conditions that can coexist with other special conditions.
export type StackingStatus = "poison" | "burn";
export type SpecialStatus = NonStackingStatus | StackingStatus;

export type Side = "me" | "opponent";

export type Zone = "active" | "bench";

// PokemonRef stores only the data needed by the companion app once a card is selected.
export interface PokemonRef {
  id: number;
  name: string;
}

// SlotState models a single active or bench slot on the board.
export interface SlotState {
  pokemon: PokemonRef | null;
  damage: number;
  statuses: SpecialStatus[];
}

// PlayerState keeps all per-player board data and once-per-game markers.
export interface PlayerState {
  active: SlotState;
  bench: [SlotState, SlotState, SlotState, SlotState, SlotState];
  usedGX: boolean;
  usedVSTAR: boolean;
}

// GameState is the full persisted match snapshot shared by the app.
export interface GameState {
  me: PlayerState;
  opponent: PlayerState;
  coinResult: "Heads" | "Tails" | null;
  dieResult: number | null;
  coinFlipping: boolean;
}

// PokemonSearchResult mirrors the backend autocomplete payload.
export interface PokemonSearchResult {
  id: number;
  name: string;
}
