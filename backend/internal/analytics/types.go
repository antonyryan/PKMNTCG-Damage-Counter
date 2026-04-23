package analytics

import "time"

type PokemonUsageEntry struct {
	PokemonID int    `json:"pokemonId"`
	Name      string `json:"name"`
	UseCount  int64  `json:"useCount"`
}

type DamageTotalsResponse struct {
	TotalDealt  int64 `json:"totalDealt"`
	TotalHealed int64 `json:"totalHealed"`
}

type KnockoutTotalResponse struct {
	TotalKnockouts int64 `json:"totalKnockouts"`
}

type NameResolver interface {
	NameByID(id int) (string, bool)
}

type ActionLike interface {
	GetType() string
	GetPokemonID() *int
	GetAmount() *int
	GetSide() string
	GetZone() string
	GetBenchIndex() *int
}

type RecordEvent struct {
	SessionID string
	Action    ActionLike
	Now       time.Time
}
