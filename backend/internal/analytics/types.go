package analytics

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

type VisitSummaryResponse struct {
	UniqueVisitors int64 `json:"uniqueVisitors"`
	TotalVisits    int64 `json:"totalVisits"`
	DAU            int64 `json:"dau"`
	MAU            int64 `json:"mau"`
}

type VisitorStatsResponse struct {
	VisitorID  string `json:"visitorId"`
	VisitCount int64  `json:"visitCount"`
	LastVisit  string `json:"lastVisit"`
}

type NameResolver interface {
	NameByID(id int) (string, bool)
}
