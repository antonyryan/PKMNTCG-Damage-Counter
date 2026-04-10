package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchPokemon_EmptyQueryReturnsUpTo10(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []Pokemon
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) != 10 {
		t.Fatalf("expected 10 pokemons, got %d", len(got))
	}
}

func TestSearchPokemon_FilterAndSort(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?q=ar", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []Pokemon
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) == 0 {
		t.Fatal("expected at least one pokemon in results")
	}

	for i := 1; i < len(got); i++ {
		if got[i-1].Name > got[i].Name {
			t.Fatalf("expected alphabetical sort, got %q before %q", got[i-1].Name, got[i].Name)
		}
	}

	for _, p := range got {
		if p.Name == "Pikachu" {
			t.Fatalf("unexpected pokemon in filtered response: %s", p.Name)
		}
	}
}

func TestSearchPokemon_NoMatch(t *testing.T) {
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?q=xyzxyz", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []Pokemon
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) != 0 {
		t.Fatalf("expected 0 pokemons, got %d", len(got))
	}
}
