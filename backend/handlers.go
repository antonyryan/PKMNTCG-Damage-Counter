package main

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxSearchResults = 10

// healthHandler provides a simple liveness endpoint for development and tooling.
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// searchPokemonHandler serves autocomplete results from the mock catalog.
func searchPokemonHandler(c *gin.Context) {
	q := strings.TrimSpace(strings.ToLower(c.Query("q")))
	results := filterPokemon(pokemonMock, q, maxSearchResults)
	c.JSON(http.StatusOK, results)
}

// filterPokemon centralizes the search behavior so it can be reused and tested in isolation.
func filterPokemon(list []Pokemon, query string, limit int) []Pokemon {
	if query == "" {
		// Return a shallow copy so handlers never expose the backing array directly.
		max := min(limit, len(list))
		return append([]Pokemon(nil), list[:max]...)
	}

	results := make([]Pokemon, 0, limit)
	for _, p := range list {
		if strings.Contains(strings.ToLower(p.Name), query) {
			results = append(results, p)
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// min avoids leaking helper logic into handlers that only need a small bound check.
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
