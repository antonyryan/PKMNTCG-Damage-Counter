package httpapi

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"pkmntcg/backend/internal/analytics"
	"pkmntcg/backend/internal/catalog"
	"pkmntcg/backend/internal/session"
)

func TestHTTPAPI_RotasContrato_ResultadoEsperado(t *testing.T) {
	catalogSvc := catalog.MustLoad()
	sessionStore := session.NewStore(filepath.Join(t.TempDir(), "sessions"))
	analyticsStore := analytics.MustNew(filepath.Join(t.TempDir(), "analytics.db"))
	t.Cleanup(func() { analyticsStore.Shutdown() })

	h := NewHandlers(catalogSvc, sessionStore, analyticsStore)
	r := SetupRouter(h)

	t.Run("Health_RotaDisponivel_DeveRetornar200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Search_RotaDisponivel_DeveRetornar200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/pokemon/search?limit=1", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("VisitTracking_PayloadValido_DeveRetornar202", func(t *testing.T) {
		payload := `{"visitorId":"11111111-2222-3333-4444-555555555555","visitedAt":"2026-04-23T15:04:05Z","source":"web"}`
		req := httptest.NewRequest(http.MethodPost, "/api/analytics/visit", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusAccepted {
			t.Fatalf("expected status 202, got %d", w.Code)
		}
	})

	t.Run("VisitTracking_PayloadInvalido_DeveRetornar400", func(t *testing.T) {
		payload := `{"visitorId":"abc","visitedAt":"invalid-date","source":"web"}`
		req := httptest.NewRequest(http.MethodPost, "/api/analytics/visit", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("VisitsSummary_AposEventos_DeveRetornar200", func(t *testing.T) {
		payload := `{"visitorId":"aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee","visitedAt":"2026-04-23T16:00:00Z","source":"web"}`
		postReq := httptest.NewRequest(http.MethodPost, "/api/analytics/visit", bytes.NewBufferString(payload))
		postReq.Header.Set("Content-Type", "application/json")
		postW := httptest.NewRecorder()
		r.ServeHTTP(postW, postReq)
		if postW.Code != http.StatusAccepted {
			t.Fatalf("expected status 202, got %d", postW.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/analytics/visits/summary", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("VisitorStats_VisitanteExistente_DeveRetornar200", func(t *testing.T) {
		visitorID := "99999999-8888-7777-6666-555555555555"
		payload := `{"visitorId":"` + visitorID + `","visitedAt":"2026-04-23T17:00:00Z","source":"web"}`
		postReq := httptest.NewRequest(http.MethodPost, "/api/analytics/visit", bytes.NewBufferString(payload))
		postReq.Header.Set("Content-Type", "application/json")
		postW := httptest.NewRecorder()
		r.ServeHTTP(postW, postReq)
		if postW.Code != http.StatusAccepted {
			t.Fatalf("expected status 202, got %d", postW.Code)
		}

		req := httptest.NewRequest(http.MethodGet, "/api/analytics/visits/"+visitorID, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})
}
