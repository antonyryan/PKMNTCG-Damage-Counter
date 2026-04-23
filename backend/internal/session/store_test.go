package session

import (
	"os"
	"testing"
)

func TestSessionStore_Persistencia_ResultadoEsperado(t *testing.T) {
	store := NewStore(t.TempDir())
	catalog := fakeCatalog{}

	t.Run("GetOrCreate_SemID_DeveCriarSessao", func(t *testing.T) {
		s, err := store.GetOrCreate("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if s.SessionID == "" {
			t.Fatal("expected non-empty session id")
		}
		t.Logf("session criada: %s", s.SessionID)
	})

	t.Run("ApplyAction_AposPersistir_DeveRecuperarHistorico", func(t *testing.T) {
		s, err := store.GetOrCreate("")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = store.ApplyAction(s.SessionID, ActionRequest{Type: ActionSetPokemon, Side: SideMe, Zone: ZoneActive, PokemonID: intPtr(1)}, catalog)
		if err != nil {
			t.Fatalf("unexpected error applying action: %v", err)
		}
		loaded, err := store.Get(s.SessionID)
		if err != nil {
			t.Fatalf("unexpected error loading session: %v", err)
		}
		if len(loaded.History) != 1 {
			t.Fatalf("expected history len=1, got %d", len(loaded.History))
		}
	})

	t.Run("Get_SessaoInexistente_DeveRetornarNotExist", func(t *testing.T) {
		_, err := store.Get("sessao-inexistente")
		if err == nil || !os.IsNotExist(err) {
			t.Fatalf("expected os.ErrNotExist, got %v", err)
		}
	})
}
