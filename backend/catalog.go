package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const defaultSearchLimit = 20

// PokemonCatalog owns the authoritative Pokemon dataset and evolution graph.
type PokemonCatalog struct {
	ordered []Pokemon
	byID    map[int]PokemonCatalogEntry
}

type pokemonCatalogRawEntry struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	EvolvesTo   []string `json:"evolvesTo"`
	EvolvesFrom *string  `json:"evolvesFrom"`
}

func mustLoadPokemonCatalog() *PokemonCatalog {
	catalog, err := loadPokemonCatalog()
	if err != nil {
		panic(err)
	}
	return catalog
}

func loadPokemonCatalog() (*PokemonCatalog, error) {
	_, currentFile, _, _ := runtime.Caller(0)
	backendDir := filepath.Dir(currentFile)
	paths := []string{
		filepath.Join(backendDir, "..", "pokemon_data.json"),
		filepath.Join(backendDir, "pokemon_data.json"),
		filepath.Join("..", "pokemon_data.json"),
		"pokemon_data.json",
	}

	var raw []byte
	var readErr error
	for _, path := range paths {
		raw, readErr = os.ReadFile(path)
		if readErr == nil {
			break
		}
	}

	if readErr != nil {
		remoteURL := os.Getenv("POKEMON_DATA_URL")
		if remoteURL == "" {
			remoteURL = "https://raw.githubusercontent.com/antonyryan/PKMNTCG-Damage-Counter/master/pokemon_data.json"
		}

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Get(remoteURL)
		if err != nil {
			return nil, fmt.Errorf("read pokemon catalog: %w", readErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("read pokemon catalog: %w", readErr)
		}

		raw, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("read pokemon catalog: %w", readErr)
		}
	}

	var parsed []pokemonCatalogRawEntry
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("parse pokemon catalog: %w", err)
	}

	catalog := &PokemonCatalog{
		ordered: make([]Pokemon, 0, len(parsed)),
		byID:    make(map[int]PokemonCatalogEntry, len(parsed)),
	}

	for _, entry := range parsed {
		id, err := parsePokemonID(entry.ID)
		if err != nil {
			return nil, err
		}

		evolvesTo := make([]int, 0, len(entry.EvolvesTo))
		for _, targetID := range entry.EvolvesTo {
			parsedID, err := parsePokemonID(targetID)
			if err != nil {
				return nil, err
			}
			evolvesTo = append(evolvesTo, parsedID)
		}

		var evolvesFrom *int
		if entry.EvolvesFrom != nil {
			parsedID, err := parsePokemonID(*entry.EvolvesFrom)
			if err != nil {
				return nil, err
			}
			evolvesFrom = &parsedID
		}

		catalog.ordered = append(catalog.ordered, Pokemon{ID: id, Name: entry.Name})
		catalog.byID[id] = PokemonCatalogEntry{
			Pokemon:     Pokemon{ID: id, Name: entry.Name},
			EvolvesTo:   evolvesTo,
			EvolvesFrom: evolvesFrom,
		}
	}

	sort.Slice(catalog.ordered, func(i, j int) bool {
		return catalog.ordered[i].ID < catalog.ordered[j].ID
	})

	return catalog, nil
}

func parsePokemonID(raw string) (int, error) {
	var id int
	_, err := fmt.Sscanf(raw, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("parse pokemon id %q: %w", raw, err)
	}
	return id, nil
}

func (catalog *PokemonCatalog) Search(query string, limit int) []Pokemon {
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	normalizedQuery := strings.TrimSpace(strings.ToLower(query))
	results := make([]Pokemon, 0, limit)
	for _, pokemon := range catalog.ordered {
		if normalizedQuery != "" && !strings.Contains(strings.ToLower(pokemon.Name), normalizedQuery) {
			continue
		}
		results = append(results, pokemon)
		if len(results) >= limit {
			break
		}
	}
	return results
}

func (catalog *PokemonCatalog) Get(id int) (PokemonCatalogEntry, bool) {
	pokemon, ok := catalog.byID[id]
	return pokemon, ok
}

func (catalog *PokemonCatalog) EvolutionOptions(currentPokemonID int, query string) ([]EvolutionOption, error) {
	current, ok := catalog.byID[currentPokemonID]
	if !ok {
		return nil, errors.New("pokemon not found")
	}

	rootID := current.Pokemon.ID
	for current.EvolvesFrom != nil {
		previous, ok := catalog.byID[*current.EvolvesFrom]
		if !ok {
			break
		}
		current = previous
		rootID = previous.Pokemon.ID
	}

	depthByID := make(map[int]int)
	queue := []struct {
		id    int
		depth int
	}{{id: rootID, depth: 0}}

	for len(queue) > 0 {
		next := queue[0]
		queue = queue[1:]
		if _, seen := depthByID[next.id]; seen {
			continue
		}
		depthByID[next.id] = next.depth
		pokemon, ok := catalog.byID[next.id]
		if !ok {
			continue
		}
		for _, evolveID := range pokemon.EvolvesTo {
			queue = append(queue, struct {
				id    int
				depth int
			}{id: evolveID, depth: next.depth + 1})
		}
	}

	candidateIDs := make(map[int]struct{})
	visited := make(map[int]struct{})
	traversal := []int{currentPokemonID}

	for len(traversal) > 0 {
		nextID := traversal[0]
		traversal = traversal[1:]
		if _, seen := visited[nextID]; seen {
			continue
		}
		visited[nextID] = struct{}{}

		pokemon, ok := catalog.byID[nextID]
		if !ok {
			continue
		}

		for _, evolveID := range pokemon.EvolvesTo {
			candidateIDs[evolveID] = struct{}{}
			traversal = append(traversal, evolveID)
		}
		if pokemon.EvolvesFrom != nil {
			candidateIDs[*pokemon.EvolvesFrom] = struct{}{}
			traversal = append(traversal, *pokemon.EvolvesFrom)
		}
	}

	delete(candidateIDs, currentPokemonID)
	currentDepth := depthByID[currentPokemonID]
	normalizedQuery := strings.TrimSpace(strings.ToLower(query))
	results := make([]EvolutionOption, 0, len(candidateIDs))
	for candidateID := range candidateIDs {
		candidate, ok := catalog.byID[candidateID]
		if !ok {
			continue
		}
		if normalizedQuery != "" && !strings.Contains(strings.ToLower(candidate.Pokemon.Name), normalizedQuery) {
			continue
		}
		action := "De-evolve"
		if depthByID[candidateID] >= currentDepth {
			action = "Evolve"
		}
		results = append(results, EvolutionOption{
			ID:     candidate.Pokemon.ID,
			Name:   candidate.Pokemon.Name,
			Action: action,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	return results, nil
}

func (catalog *PokemonCatalog) IsValidEvolutionTarget(currentPokemonID, targetPokemonID int) bool {
	options, err := catalog.EvolutionOptions(currentPokemonID, "")
	if err != nil {
		return false
	}
	for _, option := range options {
		if option.ID == targetPokemonID {
			return true
		}
	}
	return false
}
