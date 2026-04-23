package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
)

type CatalogLookup interface {
	FindByID(id int) (PokemonInfo, bool)
	IsValidEvolutionTarget(currentPokemonID, targetPokemonID int) bool
}

type PokemonInfo struct {
	ID   int
	Name string
}

func EmptySlot() SlotState {
	return SlotState{Pokemon: nil, Damage: 0, Statuses: []SpecialStatus{}}
}

func emptyPlayer() PlayerState {
	return PlayerState{Active: EmptySlot(), Bench: [5]SlotState{EmptySlot(), EmptySlot(), EmptySlot(), EmptySlot(), EmptySlot()}}
}

func InitialState() GameState {
	return GameState{Me: emptyPlayer(), Opponent: emptyPlayer(), CoinResult: nil, DieResult: nil, CoinFlipping: false}
}

func NormalizeGameState(state *GameState) {
	normalizePlayerState(&state.Me)
	normalizePlayerState(&state.Opponent)
}

func normalizeSlot(slot *SlotState) {
	if slot.Statuses == nil {
		slot.Statuses = []SpecialStatus{}
	}
}

func normalizePlayerState(player *PlayerState) {
	normalizeSlot(&player.Active)
	for i := range player.Bench {
		normalizeSlot(&player.Bench[i])
	}
}

func computeDamage(currentDamage, amount int) int {
	next := currentDamage + amount
	if next < 0 {
		return 0
	}
	return next
}

func computeNextStatuses(current []SpecialStatus, status SpecialStatus) []SpecialStatus {
	has := false
	for _, existing := range current {
		if existing == status {
			has = true
			break
		}
	}

	if _, exclusive := exclusiveStatuses[status]; exclusive {
		withoutExclusive := make([]SpecialStatus, 0, len(current))
		for _, existing := range current {
			if _, isExclusive := exclusiveStatuses[existing]; !isExclusive {
				withoutExclusive = append(withoutExclusive, existing)
			}
		}
		if has {
			return withoutExclusive
		}
		return append(withoutExclusive, status)
	}

	if has {
		next := make([]SpecialStatus, 0, len(current)-1)
		for _, existing := range current {
			if existing != status {
				next = append(next, existing)
			}
		}
		return next
	}

	next := append([]SpecialStatus(nil), current...)
	return append(next, status)
}

func promoteBenchWithSwap(player PlayerState, benchIndex int) PlayerState {
	if benchIndex < 0 || benchIndex >= len(player.Bench) {
		return player
	}
	fromBench := player.Bench[benchIndex]
	if fromBench.Pokemon == nil {
		return player
	}

	nextBench := player.Bench
	nextBench[benchIndex] = player.Active
	nextBench[benchIndex].Statuses = []SpecialStatus{}

	nextActive := fromBench
	nextActive.Statuses = []SpecialStatus{}

	player.Active = nextActive
	player.Bench = nextBench
	return player
}

func getPlayer(state *GameState, side Side) (*PlayerState, error) {
	switch side {
	case SideMe:
		return &state.Me, nil
	case SideOpponent:
		return &state.Opponent, nil
	default:
		return nil, fmt.Errorf("invalid side %q", side)
	}
}

func getSlot(player *PlayerState, zone Zone, benchIndex *int) (*SlotState, error) {
	switch zone {
	case ZoneActive:
		return &player.Active, nil
	case ZoneBench:
		if benchIndex == nil || *benchIndex < 0 || *benchIndex >= len(player.Bench) {
			return nil, fmt.Errorf("invalid benchIndex")
		}
		return &player.Bench[*benchIndex], nil
	default:
		return nil, fmt.Errorf("invalid zone %q", zone)
	}
}

func ApplyAction(state *GameState, req ActionRequest, catalog CatalogLookup) error {
	switch req.Type {
	case ActionSetPokemon:
		return applySetPokemon(state, req, catalog, false)
	case ActionEvolve:
		return applySetPokemon(state, req, catalog, true)
	case ActionKnockout:
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		slot, err := getSlot(player, req.Zone, req.BenchIndex)
		if err != nil {
			return err
		}
		*slot = EmptySlot()
		return nil
	case ActionAdjust:
		if req.Amount == nil {
			return fmt.Errorf("amount is required")
		}
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		slot, err := getSlot(player, req.Zone, req.BenchIndex)
		if err != nil {
			return err
		}
		slot.Damage = computeDamage(slot.Damage, *req.Amount)
		return nil
	case ActionStatus:
		if req.Status == nil {
			return fmt.Errorf("status is required")
		}
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		player.Active.Statuses = computeNextStatuses(player.Active.Statuses, *req.Status)
		return nil
	case ActionPromote:
		if req.BenchIndex == nil {
			return fmt.Errorf("benchIndex is required")
		}
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		*player = promoteBenchWithSwap(*player, *req.BenchIndex)
		return nil
	case ActionToggleGX:
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		player.UsedGX = !player.UsedGX
		return nil
	case ActionToggleVSTAR:
		player, err := getPlayer(state, req.Side)
		if err != nil {
			return err
		}
		player.UsedVSTAR = !player.UsedVSTAR
		return nil
	case ActionFlipCoin:
		result := "Tails"
		if randomBool() {
			result = "Heads"
		}
		state.CoinFlipping = false
		state.CoinResult = &result
		return nil
	case ActionRollDie:
		result := randomDieRoll()
		state.DieResult = &result
		return nil
	case ActionReset:
		*state = InitialState()
		return nil
	default:
		return fmt.Errorf("unsupported action %q", req.Type)
	}
}

func applySetPokemon(state *GameState, req ActionRequest, catalog CatalogLookup, validateEvolution bool) error {
	if req.PokemonID == nil {
		return fmt.Errorf("pokemonId is required")
	}

	catalogPokemon, ok := catalog.FindByID(*req.PokemonID)
	if !ok {
		return fmt.Errorf("pokemon %d not found", *req.PokemonID)
	}

	player, err := getPlayer(state, req.Side)
	if err != nil {
		return err
	}
	if req.Zone == "" {
		return fmt.Errorf("zone is required")
	}
	if validateEvolution {
		slot, err := getSlot(player, req.Zone, req.BenchIndex)
		if err != nil {
			return err
		}
		if slot.Pokemon == nil {
			return fmt.Errorf("cannot evolve empty slot")
		}
		if !catalog.IsValidEvolutionTarget(slot.Pokemon.ID, *req.PokemonID) {
			return fmt.Errorf("pokemon %d cannot evolve to %d", slot.Pokemon.ID, *req.PokemonID)
		}
	}

	slot, err := getSlot(player, req.Zone, req.BenchIndex)
	if err != nil {
		return err
	}
	hadPokemon := slot.Pokemon != nil
	next := SlotState{Pokemon: &PokemonRef{ID: catalogPokemon.ID, Name: catalogPokemon.Name}, Damage: 0}
	if hadPokemon {
		next.Damage = slot.Damage
	}
	if req.Zone == ZoneActive {
		next.Statuses = append([]SpecialStatus(nil), slot.Statuses...)
	} else {
		next.Statuses = []SpecialStatus{}
	}
	*slot = next
	return nil
}

func GenerateSessionID() (string, error) {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return hex.EncodeToString(raw), nil
}

func randomBool() bool {
	n, err := rand.Int(rand.Reader, big.NewInt(2))
	if err != nil {
		return false
	}
	return n.Int64() == 1
}

func randomDieRoll() int {
	n, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return 1
	}
	return int(n.Int64()) + 1
}
