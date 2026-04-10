import type {
  GameHistoryEntry,
  GameState,
  PokemonEvolutionOption,
  PokemonRef,
  PokemonSearchResult,
  Side,
  SpecialStatus,
  Zone,
} from "../types";

export interface SessionSnapshotResponse {
  sessionId: string;
  state: GameState;
  history: GameHistoryEntry[];
  updatedAt: string;
}

export interface SessionActionRequest {
  type:
    | "set-pokemon"
    | "evolve-pokemon"
    | "knockout"
    | "adjust-damage"
    | "toggle-status"
    | "promote-bench"
    | "toggle-gx"
    | "toggle-vstar"
    | "flip-coin"
    | "roll-die"
    | "reset";
  side?: Side;
  zone?: Zone;
  benchIndex?: number;
  pokemonId?: number;
  amount?: number;
  status?: SpecialStatus;
}

export function buildApiUrl(
  path: string,
  query?: Record<string, string | number | undefined>,
): string {
  const baseOrigin = typeof window === "undefined"
    ? "http://localhost"
    : window.location.origin;
  const url = new URL(path, baseOrigin);
  if (query) {
    for (const [key, value] of Object.entries(query)) {
      if (value === undefined || value === "") {
        continue;
      }
      url.searchParams.set(key, String(value));
    }
  }
  return `${url.pathname}${url.search}`;
}

async function fetchJson<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers ?? {}),
    },
    ...init,
  });

  if (!response.ok) {
    const message = await response.text();
    throw new Error(message || `Request failed with status ${response.status}`);
  }

  return (await response.json()) as T;
}

export function searchPokemon(
  query: string,
  limit = 20,
): Promise<PokemonSearchResult[]> {
  return fetchJson<PokemonSearchResult[]>(
    buildApiUrl("/api/pokemon/search", { q: query, limit }),
  );
}

export function getEvolutionOptions(
  pokemonId: number,
  query: string,
): Promise<PokemonEvolutionOption[]> {
  return fetchJson<PokemonEvolutionOption[]>(
    buildApiUrl(`/api/pokemon/${pokemonId}/evolution-options`, { q: query }),
  );
}

export function createOrLoadSession(
  sessionId: string | null,
): Promise<SessionSnapshotResponse> {
  return fetchJson<SessionSnapshotResponse>("/api/sessions", {
    method: "POST",
    body: JSON.stringify(sessionId ? { sessionId } : {}),
  });
}

export function applySessionAction(
  sessionId: string,
  action: SessionActionRequest,
): Promise<SessionSnapshotResponse> {
  return fetchJson<SessionSnapshotResponse>(`/api/sessions/${sessionId}/actions`, {
    method: "POST",
    body: JSON.stringify(action),
  });
}

export function toActionPokemon(pokemon: PokemonRef): number {
  return pokemon.id;
}
