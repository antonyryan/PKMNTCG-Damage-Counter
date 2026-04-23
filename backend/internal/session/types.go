package session

import "time"

type SpecialStatus string
type Side string
type Zone string

const (
	StatusSleep     SpecialStatus = "sleep"
	StatusConfusion SpecialStatus = "confusion"
	StatusParalysis SpecialStatus = "paralysis"
	StatusPoison    SpecialStatus = "poison"
	StatusBurn      SpecialStatus = "burn"

	SideMe       Side = "me"
	SideOpponent Side = "opponent"

	ZoneActive Zone = "active"
	ZoneBench  Zone = "bench"
)

const (
	ActionSetPokemon  = "set-pokemon"
	ActionEvolve      = "evolve-pokemon"
	ActionKnockout    = "knockout"
	ActionAdjust      = "adjust-damage"
	ActionStatus      = "toggle-status"
	ActionPromote     = "promote-bench"
	ActionToggleGX    = "toggle-gx"
	ActionToggleVSTAR = "toggle-vstar"
	ActionFlipCoin    = "flip-coin"
	ActionRollDie     = "roll-die"
	ActionReset       = "reset"
)

var exclusiveStatuses = map[SpecialStatus]struct{}{
	StatusSleep:     {},
	StatusConfusion: {},
	StatusParalysis: {},
}

type PokemonRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type SlotState struct {
	Pokemon  *PokemonRef     `json:"pokemon"`
	Damage   int             `json:"damage"`
	Statuses []SpecialStatus `json:"statuses"`
}

type PlayerState struct {
	Active    SlotState    `json:"active"`
	Bench     [5]SlotState `json:"bench"`
	UsedGX    bool         `json:"usedGX"`
	UsedVSTAR bool         `json:"usedVSTAR"`
}

type GameState struct {
	Me           PlayerState `json:"me"`
	Opponent     PlayerState `json:"opponent"`
	CoinResult   *string     `json:"coinResult"`
	DieResult    *int        `json:"dieResult"`
	CoinFlipping bool        `json:"coinFlipping"`
}

type ActionRequest struct {
	Type       string         `json:"type"`
	Side       Side           `json:"side,omitempty"`
	Zone       Zone           `json:"zone,omitempty"`
	BenchIndex *int           `json:"benchIndex,omitempty"`
	PokemonID  *int           `json:"pokemonId,omitempty"`
	Amount     *int           `json:"amount,omitempty"`
	Status     *SpecialStatus `json:"status,omitempty"`
}

type HistoryEntry struct {
	Timestamp  time.Time      `json:"timestamp"`
	Action     string         `json:"action"`
	Side       Side           `json:"side,omitempty"`
	Zone       Zone           `json:"zone,omitempty"`
	BenchIndex *int           `json:"benchIndex,omitempty"`
	PokemonID  *int           `json:"pokemonId,omitempty"`
	Amount     *int           `json:"amount,omitempty"`
	Status     *SpecialStatus `json:"status,omitempty"`
}

type GameSession struct {
	SessionID string         `json:"sessionId"`
	State     GameState      `json:"state"`
	History   []HistoryEntry `json:"history"`
	UpdatedAt time.Time      `json:"updatedAt"`
}
