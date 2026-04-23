package analytics

import (
	"path/filepath"
	"testing"
	"time"

	"pkmntcg/backend/internal/session"
)

type staticNames struct{}

func (staticNames) NameByID(id int) (string, bool) {
	if id == 1 {
		return "Bulbasaur", true
	}
	return "", false
}

func TestAnalyticsStore_WorkerChannel_ResultadoEsperado(t *testing.T) {
	store := MustNew(filepath.Join(t.TempDir(), "analytics.db"))
	t.Cleanup(func() { store.Shutdown() })

	t.Run("Record_SetPokemon_DeveIncrementarUso", func(t *testing.T) {
		store.Record("s1", session.ActionRequest{Type: session.ActionSetPokemon, PokemonID: intPtr(1)}, staticNames{}, time.Now())
		store.Record("s1", session.ActionRequest{Type: session.ActionSetPokemon, PokemonID: intPtr(1)}, staticNames{}, time.Now())
		entries, err := store.QueryTopPokemon(10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(entries) == 0 || entries[0].PokemonID != 1 || entries[0].UseCount != 2 {
			t.Fatalf("unexpected entries: %+v", entries)
		}
	})

	t.Run("Record_Knockout_DeveSomarTotal", func(t *testing.T) {
		store.Record("s1", session.ActionRequest{Type: session.ActionKnockout}, staticNames{}, time.Now())
		totals, err := store.QueryKnockouts()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if totals.TotalKnockouts < 1 {
			t.Fatalf("expected knockout total >= 1, got %d", totals.TotalKnockouts)
		}
	})
}

func intPtr(v int) *int { return &v }
