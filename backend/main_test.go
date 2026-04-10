package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func createTestSession(t *testing.T, router http.Handler) GameSession {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var session GameSession
	if err := json.Unmarshal(w.Body.Bytes(), &session); err != nil {
		t.Fatalf("failed to parse session response: %v", err)
	}

	return session
}

func applyActionRequest(t *testing.T, router http.Handler, sessionID string, payload string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/sessions/"+sessionID+"/actions", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestSearchPokemon_EmptyQueryReturnsCatalogInIDOrder(t *testing.T) {
	sessionStore = newSessionStore(t.TempDir())
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?limit=3", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []Pokemon
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) != 3 {
		t.Fatalf("expected 3 pokemons, got %d", len(got))
	}

	if got[0].ID != 1 || got[1].ID != 2 || got[2].ID != 3 {
		t.Fatalf("expected pokemon ids [1 2 3], got [%d %d %d]", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestSearchPokemon_FilterBySubstring(t *testing.T) {
	sessionStore = newSessionStore(t.TempDir())
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?q=saur&limit=5", nil)
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

	for _, p := range got {
		if p.Name != "Bulbasaur" && p.Name != "Ivysaur" && p.Name != "Venusaur" {
			t.Fatalf("unexpected pokemon in filtered response: %s", p.Name)
		}
	}
}

func TestEvolutionOptions_ReturnsFullChainAndActions(t *testing.T) {
	sessionStore = newSessionStore(t.TempDir())
	r := setupRouter()
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/1/evolution-options", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []EvolutionOption
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 evolution options, got %d", len(got))
	}

	if got[0].ID != 2 || got[0].Action != "Evolve" {
		t.Fatalf("expected first option to be id=2 action=Evolve, got id=%d action=%s", got[0].ID, got[0].Action)
	}
	if got[1].ID != 3 || got[1].Action != "Evolve" {
		t.Fatalf("expected second option to be id=3 action=Evolve, got id=%d action=%s", got[1].ID, got[1].Action)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/pokemon/3/evolution-options", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 de-evolution options, got %d", len(got))
	}
	if got[0].Action != "De-evolve" || got[1].Action != "De-evolve" {
		t.Fatalf("expected de-evolve actions, got %+v", got)
	}
	if got[0].ID != 1 || got[1].ID != 2 {
		t.Fatalf("expected ids [1 2], got [%d %d]", got[0].ID, got[1].ID)
	}
}

func TestSessionActions_ValidateEvolutionTransitionAndPersistHistory(t *testing.T) {
	sessionStore = newSessionStore(t.TempDir())
	r := setupRouter()
	session := createTestSession(t, r)

	w := applyActionRequest(t, r, session.SessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	w = applyActionRequest(t, r, session.SessionID, `{"type":"evolve-pokemon","side":"me","zone":"active","pokemonId":3}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got GameSession
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse session response: %v", err)
	}

	if got.State.Me.Active.Pokemon == nil || got.State.Me.Active.Pokemon.ID != 3 {
		t.Fatalf("expected active pokemon id 3 after evolve, got %+v", got.State.Me.Active.Pokemon)
	}
	if len(got.History) != 2 {
		t.Fatalf("expected 2 history events, got %d", len(got.History))
	}

	w = applyActionRequest(t, r, session.SessionID, `{"type":"evolve-pokemon","side":"me","zone":"active","pokemonId":25}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
