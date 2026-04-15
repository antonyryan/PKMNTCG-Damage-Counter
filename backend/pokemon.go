package main

// Pokemon is the minimal response shape used by autocomplete and board state payloads.
type Pokemon struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// PokemonCatalogEntry is the backend-only representation enriched with evolution links.
type PokemonCatalogEntry struct {
	Pokemon     Pokemon
	EvolvesTo   []int
	EvolvesFrom *int
}

// EvolutionOption is returned by the evolution-options endpoint.
type EvolutionOption struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Action string `json:"action"`
}
