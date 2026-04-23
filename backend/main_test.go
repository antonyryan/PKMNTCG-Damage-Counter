package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"pkmntcg/backend/internal/app"
)

type testServer struct {
	handler   http.Handler
	analytics interface{ Shutdown() }
}

func newTestServer(t *testing.T) testServer {
	t.Helper()
	instance := app.New("0", t.TempDir())
	t.Cleanup(func() { instance.Analytics.Shutdown() })
	return testServer{handler: instance.Server.Handler, analytics: instance.Analytics}
}

func createTestSession(t *testing.T, router http.Handler) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var session map[string]any
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

func getSessionID(t *testing.T, session map[string]any) string {
	t.Helper()
	id, ok := session["sessionId"].(string)
	if !ok || id == "" {
		t.Fatalf("invalid sessionId in response: %+v", session)
	}
	return id
}

func TestIntegracaoCatalog_SearchQueryVazia_RetornaOrdemPorID(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?limit=3", nil)
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 pokemons, got %d", len(got))
	}

	id0 := int(got[0]["id"].(float64))
	id1 := int(got[1]["id"].(float64))
	id2 := int(got[2]["id"].(float64))
	if id0 != 1 || id1 != 2 || id2 != 3 {
		t.Fatalf("expected pokemon ids [1 2 3], got [%d %d %d]", id0, id1, id2)
	}
}

func TestIntegracaoCatalog_SearchComFiltro_RetornaSomenteCorrespondencias(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?q=saur&limit=5", nil)
	w := httptest.NewRecorder()

	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got []map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected at least one pokemon in results")
	}

	for _, p := range got {
		name := p["name"].(string)
		if name != "Bulbasaur" && name != "Ivysaur" && name != "Venusaur" {
			t.Fatalf("unexpected pokemon in filtered response: %s", name)
		}
	}
}

func TestIntegracaoCatalog_EvolutionOptionsComCadeiaCompleta_RetornaAcoesValidas(t *testing.T) {
	s := newTestServer(t)
	t.Run("Bulbasaur_DeveRetornarIvysaurEVenusaurComoEvolve", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/pokemon/1/evolution-options", nil)
		w := httptest.NewRecorder()

		s.handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		var got []map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("expected 2 evolution options, got %d", len(got))
		}

		if int(got[0]["id"].(float64)) != 2 || got[0]["action"].(string) != "Evolve" {
			t.Fatalf("unexpected first option: %+v", got[0])
		}
		if int(got[1]["id"].(float64)) != 3 || got[1]["action"].(string) != "Evolve" {
			t.Fatalf("unexpected second option: %+v", got[1])
		}
	})
}

func TestIntegracaoSession_AcaoEvolucaoValida_PersisteHistorico(t *testing.T) {
	s := newTestServer(t)
	session := createTestSession(t, s.handler)
	sessionID := getSessionID(t, session)

	w := applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	w = applyActionRequest(t, s.handler, sessionID, `{"type":"evolve-pokemon","side":"me","zone":"active","pokemonId":3}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to parse session response: %v", err)
	}

	state := got["state"].(map[string]any)
	me := state["me"].(map[string]any)
	active := me["active"].(map[string]any)
	pokemon := active["pokemon"].(map[string]any)
	if int(pokemon["id"].(float64)) != 3 {
		t.Fatalf("expected active pokemon id 3 after evolve, got %+v", pokemon)
	}

	history := got["history"].([]any)
	if len(history) != 2 {
		t.Fatalf("expected 2 history events, got %d", len(history))
	}

	w = applyActionRequest(t, s.handler, sessionID, `{"type":"evolve-pokemon","side":"me","zone":"active","pokemonId":25}`)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestIntegracaoAnalytics_SetPokemonRepetido_IncrementaUso(t *testing.T) {
	s := newTestServer(t)
	sessionID := getSessionID(t, createTestSession(t, s.handler))

	applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)

	// Read endpoint is synchronous and drains prior queued writes.
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analytics/pokemon-usage?limit=10", nil)
	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var body map[string][]map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("parse analytics response: %v", err)
	}
	entries := body["pokemon"]
	if len(entries) == 0 {
		t.Fatal("expected at least one usage entry")
	}
	if int(entries[0]["pokemonId"].(float64)) != 1 || int(entries[0]["useCount"].(float64)) != 2 {
		t.Fatalf("expected pokemonId=1 useCount=2, got %+v", entries[0])
	}
}

func TestIntegracaoAnalytics_DanoAcumulado_RetornaTotaisCorretos(t *testing.T) {
	s := newTestServer(t)
	sessionID := getSessionID(t, createTestSession(t, s.handler))

	applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	applyActionRequest(t, s.handler, sessionID, `{"type":"adjust-damage","side":"me","zone":"active","amount":50}`)
	applyActionRequest(t, s.handler, sessionID, `{"type":"adjust-damage","side":"me","zone":"active","amount":30}`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analytics/damage", nil)
	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var totals map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &totals); err != nil {
		t.Fatalf("parse analytics response: %v", err)
	}
	if int(totals["totalDealt"].(float64)) != 80 {
		t.Fatalf("expected totalDealt=80, got %v", totals["totalDealt"])
	}
}

func TestIntegracaoAnalytics_KnockoutRegistrado_IncrementaContador(t *testing.T) {
	s := newTestServer(t)
	sessionID := getSessionID(t, createTestSession(t, s.handler))

	applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	applyActionRequest(t, s.handler, sessionID, `{"type":"knockout","side":"me","zone":"active"}`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/analytics/knockouts", nil)
	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var totals map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &totals); err != nil {
		t.Fatalf("parse analytics response: %v", err)
	}
	if int(totals["totalKnockouts"].(float64)) != 1 {
		t.Fatalf("expected totalKnockouts=1, got %v", totals["totalKnockouts"])
	}
}

func TestIntegracaoAnalytics_AcoesConcorrentes_MantemConsistencia(t *testing.T) {
	s := newTestServer(t)
	const goroutines = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			sessionID := getSessionID(t, createTestSession(t, s.handler))
			applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":4}`)
			applyActionRequest(t, s.handler, sessionID, `{"type":"adjust-damage","side":"me","zone":"active","amount":10}`)
			applyActionRequest(t, s.handler, sessionID, `{"type":"knockout","side":"me","zone":"active"}`)
		}()
	}
	wg.Wait()

	t.Run("Knockouts_DeveRefletirTotalDeGoroutines", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/knockouts", nil)
		s.handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		var knockouts map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &knockouts); err != nil {
			t.Fatalf("parse knockouts response: %v", err)
		}
		if int(knockouts["totalKnockouts"].(float64)) != goroutines {
			t.Fatalf("expected %d knockouts, got %v", goroutines, knockouts["totalKnockouts"])
		}
	})

	t.Run("Usage_DeveRefletirTotalDeGoroutines", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/analytics/pokemon-usage?limit=1", nil)
		s.handler.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		var usage map[string][]map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &usage); err != nil {
			t.Fatalf("parse usage response: %v", err)
		}
		if len(usage["pokemon"]) == 0 {
			t.Fatal("expected usage response entries")
		}
		if int(usage["pokemon"][0]["useCount"].(float64)) != goroutines {
			t.Fatalf("expected useCount=%d, got %v", goroutines, usage["pokemon"][0]["useCount"])
		}
	})
}

func TestBootstrap_AplicacaoInicializada_HandlerDisponivel(t *testing.T) {
	instance := app.New("0", t.TempDir())
	defer instance.Analytics.Shutdown()

	if instance.Server == nil || instance.Server.Handler == nil {
		t.Fatal("expected bootstrapped server handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	instance.Server.Handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var health map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &health); err != nil {
		t.Fatalf("parse health response: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("expected status=ok, got %q", health["status"])
	}
}

func TestIntegracaoContratos_APISessoesEAnalytics_SemQuebraContrato(t *testing.T) {
	s := newTestServer(t)
	session := createTestSession(t, s.handler)
	sessionID := getSessionID(t, session)

	t.Logf("session criada para validação de contrato: %s", sessionID)

	w := applyActionRequest(t, s.handler, sessionID, `{"type":"set-pokemon","side":"me","zone":"active","pokemonId":1}`)
	if w.Code != http.StatusOK {
		t.Fatalf("expected set-pokemon to keep contract status %d, got %d", http.StatusOK, w.Code)
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sessionID+"/history", nil)
	s.handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected session history status %d, got %d", http.StatusOK, w.Code)
	}
}
