import rawPokemonData from "./pokemon_data.json";
import type { PokemonSearchResult } from "../types";

interface PokemonEvolutionRaw {
  id: string;
  name: string;
  evolvesTo: string[];
  evolvesFrom: string | null;
  processed?: boolean;
}

interface PokemonEvolutionNode {
  id: number;
  name: string;
  evolvesTo: number[];
  evolvesFrom: number | null;
}

const parsedPokemonData = rawPokemonData as PokemonEvolutionRaw[];

const pokemonById = new Map<number, PokemonEvolutionNode>(
  parsedPokemonData
    .map((entry) => {
      const id = Number(entry.id);
      if (!Number.isFinite(id)) {
        return null;
      }

      const evolvesTo = (entry.evolvesTo ?? [])
        .map((nextId) => Number(nextId))
        .filter((nextId) => Number.isFinite(nextId));

      const evolvesFrom =
        entry.evolvesFrom === null ? null : Number(entry.evolvesFrom);

      return [
        id,
        {
          id,
          name: entry.name,
          evolvesTo,
          evolvesFrom: Number.isFinite(evolvesFrom) ? evolvesFrom : null,
        },
      ] as const;
    })
    .filter((entry): entry is readonly [number, PokemonEvolutionNode] =>
      entry !== null,
    ),
);

const allPokemonSorted: PokemonSearchResult[] = [...pokemonById.values()]
  .sort((a, b) => a.id - b.id)
  .map((pokemon) => ({ id: pokemon.id, name: pokemon.name }));

export function searchPokemonCatalog(
  query: string,
  limit = 20,
): PokemonSearchResult[] {
  const normalizedQuery = query.trim().toLowerCase();

  const filtered = allPokemonSorted.filter((pokemon) =>
    normalizedQuery.length === 0
      ? true
      : pokemon.name.toLowerCase().includes(normalizedQuery),
  );

  return filtered.slice(0, Math.max(1, limit));
}

export function getEvolutionCandidates(
  currentPokemonId: number,
  query: string,
): PokemonSearchResult[] {
  const current = pokemonById.get(currentPokemonId);
  if (!current) {
    return [];
  }

  // Traverse both directions to collect the full evolution line so stage jumps are allowed.
  const candidateIds = new Set<number>();
  const visited = new Set<number>();
  const queue: number[] = [current.id];

  while (queue.length > 0) {
    const id = queue.shift();
    if (id === undefined || visited.has(id)) {
      continue;
    }

    visited.add(id);
    const pokemon = pokemonById.get(id);
    if (!pokemon) {
      continue;
    }

    for (const nextId of pokemon.evolvesTo) {
      candidateIds.add(nextId);
      if (!visited.has(nextId)) {
        queue.push(nextId);
      }
    }

    if (pokemon.evolvesFrom !== null) {
      candidateIds.add(pokemon.evolvesFrom);
      if (!visited.has(pokemon.evolvesFrom)) {
        queue.push(pokemon.evolvesFrom);
      }
    }
  }

  candidateIds.delete(current.id);

  const normalizedQuery = query.trim().toLowerCase();

  return [...candidateIds]
    .map((id) => pokemonById.get(id))
    .filter((pokemon): pokemon is PokemonEvolutionNode => pokemon !== undefined)
    .filter((pokemon) =>
      normalizedQuery.length === 0
        ? true
        : pokemon.name.toLowerCase().includes(normalizedQuery),
    )
    .sort((a, b) => a.name.localeCompare(b.name))
    .map((pokemon) => ({ id: pokemon.id, name: pokemon.name }));
}
