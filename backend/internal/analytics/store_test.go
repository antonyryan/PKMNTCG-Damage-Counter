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

func TestAnalyticsStore_VisitsAnonimas_ResultadoEsperado(t *testing.T) {
	store := MustNew(filepath.Join(t.TempDir(), "analytics.db"))
	t.Cleanup(func() { store.Shutdown() })

	referenceNow := time.Date(2026, 4, 23, 15, 0, 0, 0, time.UTC)
	withinDay := referenceNow.Add(-3 * time.Hour)
	withinMonth := referenceNow.Add(-10 * 24 * time.Hour)
	olderThanMonth := referenceNow.Add(-40 * 24 * time.Hour)

	store.RecordVisit("visitor-a", withinDay, "web")
	store.RecordVisit("visitor-a", withinMonth, "web")
	store.RecordVisit("visitor-b", withinMonth, "web")
	store.RecordVisit("visitor-c", olderThanMonth, "web")

	t.Run("Summary_DeveRetornarUnicosTotaisDAUeMAU", func(t *testing.T) {
		summary, err := store.QueryVisitSummary(referenceNow)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if summary.TotalVisits != 4 {
			t.Fatalf("expected totalVisits=4, got %d", summary.TotalVisits)
		}
		if summary.UniqueVisitors != 3 {
			t.Fatalf("expected uniqueVisitors=3, got %d", summary.UniqueVisitors)
		}
		if summary.DAU != 1 {
			t.Fatalf("expected dau=1, got %d", summary.DAU)
		}
		if summary.MAU != 2 {
			t.Fatalf("expected mau=2, got %d", summary.MAU)
		}
	})

	t.Run("VisitorStats_DeveRetornarRecorrenciaEUltimaVisita", func(t *testing.T) {
		stats, err := store.QueryVisitorStats("visitor-a")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats.VisitCount != 2 {
			t.Fatalf("expected visitCount=2, got %d", stats.VisitCount)
		}
		if stats.LastVisit != withinDay.UTC().Format(time.RFC3339) {
			t.Fatalf("expected lastVisit=%s, got %s", withinDay.UTC().Format(time.RFC3339), stats.LastVisit)
		}
	})
}

func intPtr(v int) *int { return &v }
