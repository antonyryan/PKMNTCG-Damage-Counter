package session

import "testing"

type fakeCatalog struct{}

func (fakeCatalog) FindByID(id int) (PokemonInfo, bool) {
	if id == 1 {
		return PokemonInfo{ID: 1, Name: "Bulbasaur"}, true
	}
	if id == 2 {
		return PokemonInfo{ID: 2, Name: "Ivysaur"}, true
	}
	return PokemonInfo{}, false
}

func (fakeCatalog) IsValidEvolutionTarget(currentPokemonID, targetPokemonID int) bool {
	return currentPokemonID == 1 && targetPokemonID == 2
}

func TestSessionDomain_AplicarAcao_ResultadoEsperado(t *testing.T) {
	t.Run("SetPokemon_NoSlotActive_DeveAdicionarPokemon", func(t *testing.T) {
		state := InitialState()
		req := ActionRequest{Type: ActionSetPokemon, Side: SideMe, Zone: ZoneActive, PokemonID: intPtr(1)}
		if err := ApplyAction(&state, req, fakeCatalog{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Me.Active.Pokemon == nil || state.Me.Active.Pokemon.ID != 1 {
			t.Fatalf("expected active pokemon id=1, got %+v", state.Me.Active.Pokemon)
		}
	})

	t.Run("Evolve_ComEvolucaoValida_DeveAtualizarPokemon", func(t *testing.T) {
		state := InitialState()
		state.Me.Active.Pokemon = &PokemonRef{ID: 1, Name: "Bulbasaur"}
		req := ActionRequest{Type: ActionEvolve, Side: SideMe, Zone: ZoneActive, PokemonID: intPtr(2)}
		if err := ApplyAction(&state, req, fakeCatalog{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Me.Active.Pokemon == nil || state.Me.Active.Pokemon.ID != 2 {
			t.Fatalf("expected active pokemon id=2, got %+v", state.Me.Active.Pokemon)
		}
	})

	t.Run("Adjust_ComDanoNegativoExcessivo_DeveLimitarEmZero", func(t *testing.T) {
		state := InitialState()
		state.Me.Active.Pokemon = &PokemonRef{ID: 1, Name: "Bulbasaur"}
		state.Me.Active.Damage = 20
		req := ActionRequest{Type: ActionAdjust, Side: SideMe, Zone: ZoneActive, Amount: intPtr(-999)}
		if err := ApplyAction(&state, req, fakeCatalog{}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if state.Me.Active.Damage != 0 {
			t.Fatalf("expected damage=0, got %d", state.Me.Active.Damage)
		}
	})
}

func intPtr(v int) *int { return &v }
