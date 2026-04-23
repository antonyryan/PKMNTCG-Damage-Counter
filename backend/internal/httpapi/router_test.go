package httpapi

import (
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
}
