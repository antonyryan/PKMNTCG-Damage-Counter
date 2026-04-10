package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "modernc.org/sqlite"
)

// healGracePeriod is the minimum time between a positive and negative damage event
// on the same slot for the negative event to be counted as intentional healing.
// Removals within this window are treated as corrections and ignored.
const healGracePeriod = 2 * time.Second

// AnalyticsStore persists gameplay metrics in SQLite.
type AnalyticsStore struct {
	db *sql.DB
}

// PokemonUsageEntry is one row returned by the top-pokemon endpoint.
type PokemonUsageEntry struct {
	PokemonID int    `json:"pokemonId"`
	Name      string `json:"name"`
	UseCount  int64  `json:"useCount"`
}

// DamageTotalsResponse aggregates damage dealt and intentionally healed across all sessions.
type DamageTotalsResponse struct {
	TotalDealt  int64 `json:"totalDealt"`
	TotalHealed int64 `json:"totalHealed"`
}

// KnockoutTotalResponse counts knockouts across all sessions.
type KnockoutTotalResponse struct {
	TotalKnockouts int64 `json:"totalKnockouts"`
}

const analyticsSchema = `
CREATE TABLE IF NOT EXISTS pokemon_usage (
	pokemon_id INTEGER PRIMARY KEY,
	name       TEXT    NOT NULL,
	use_count  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS damage_totals (
	id               INTEGER PRIMARY KEY CHECK (id = 1),
	total_dealt      INTEGER NOT NULL DEFAULT 0,
	total_healed     INTEGER NOT NULL DEFAULT 0,
	total_knockouts  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS slot_last_damage (
	session_id      TEXT    NOT NULL,
	side            TEXT    NOT NULL,
	zone            TEXT    NOT NULL,
	bench_index     INTEGER NOT NULL DEFAULT -1,
	last_positive_at INTEGER NOT NULL,
	PRIMARY KEY (session_id, side, zone, bench_index)
);

INSERT OR IGNORE INTO damage_totals (id, total_dealt, total_healed, total_knockouts)
VALUES (1, 0, 0, 0);
`

func mustNewAnalyticsStore(dbPath string) *AnalyticsStore {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("analytics: open db %q: %v", dbPath, err)
	}
	// A single writer avoids SQLITE_BUSY on concurrent action requests.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(analyticsSchema); err != nil {
		log.Fatalf("analytics: init schema: %v", err)
	}
	return &AnalyticsStore{db: db}
}

// Record inspects an already-validated-and-applied action and updates analytics counters.
// now is the wall-clock time the action was received, used for the heal-grace-period check.
func (s *AnalyticsStore) Record(sessionID string, req SessionActionRequest, catalog *PokemonCatalog, now time.Time) {
	switch req.Type {
	case ActionSetPokemon, ActionEvolve:
		if req.PokemonID == nil {
			return
		}
		name := ""
		if entry, ok := catalog.Get(*req.PokemonID); ok {
			name = entry.Pokemon.Name
		}
		s.recordPokemonUsed(*req.PokemonID, name)

	case ActionAdjust:
		if req.Amount != nil {
			s.recordDamage(sessionID, req, *req.Amount, now)
		}

	case ActionKnockout:
		s.recordKnockout()
	}
}

func (s *AnalyticsStore) recordPokemonUsed(pokemonID int, name string) {
	_, err := s.db.Exec(`
		INSERT INTO pokemon_usage (pokemon_id, name, use_count) VALUES (?, ?, 1)
		ON CONFLICT(pokemon_id) DO UPDATE SET
			use_count = use_count + 1,
			name = CASE WHEN excluded.name != '' THEN excluded.name ELSE name END
	`, pokemonID, name)
	if err != nil {
		log.Printf("analytics: recordPokemonUsed id=%d: %v", pokemonID, err)
	}
}

// slotBenchKey converts a nullable BenchIndex to a stable integer key (-1 = active slot).
func slotBenchKey(req SessionActionRequest) int {
	if req.BenchIndex == nil {
		return -1
	}
	return *req.BenchIndex
}

func (s *AnalyticsStore) recordDamage(sessionID string, req SessionActionRequest, amount int, now time.Time) {
	bk := slotBenchKey(req)
	side := string(req.Side)
	zone := string(req.Zone)

	if amount > 0 {
		// Damage dealt: increment total and refresh the slot timestamp.
		if _, err := s.db.Exec(`
			UPDATE damage_totals SET total_dealt = total_dealt + ? WHERE id = 1
		`, amount); err != nil {
			log.Printf("analytics: total_dealt += %d: %v", amount, err)
			return
		}
		if _, err := s.db.Exec(`
			INSERT INTO slot_last_damage (session_id, side, zone, bench_index, last_positive_at)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(session_id, side, zone, bench_index) DO UPDATE SET
				last_positive_at = excluded.last_positive_at
		`, sessionID, side, zone, bk, now.UnixMilli()); err != nil {
			log.Printf("analytics: slot_last_damage upsert: %v", err)
		}
		return
	}

	if amount < 0 {
		// Potential heal: only count when more than healGracePeriod has elapsed since the last
		// positive damage on this exact slot (avoids counting accidental overwrites).
		var lastMs int64
		err := s.db.QueryRow(`
			SELECT last_positive_at FROM slot_last_damage
			WHERE session_id = ? AND side = ? AND zone = ? AND bench_index = ?
		`, sessionID, side, zone, bk).Scan(&lastMs)
		if err == sql.ErrNoRows {
			// Slot never received tracked damage; removal is not a heal.
			return
		}
		if err != nil {
			log.Printf("analytics: slot_last_damage query: %v", err)
			return
		}
		if now.Sub(time.UnixMilli(lastMs)) <= healGracePeriod {
			// Within the grace period — likely a user correction, not intentional healing.
			return
		}
		healed := -amount
		if _, err := s.db.Exec(`
			UPDATE damage_totals SET total_healed = total_healed + ? WHERE id = 1
		`, healed); err != nil {
			log.Printf("analytics: total_healed += %d: %v", healed, err)
		}
	}
}

func (s *AnalyticsStore) recordKnockout() {
	if _, err := s.db.Exec(`
		UPDATE damage_totals SET total_knockouts = total_knockouts + 1 WHERE id = 1
	`); err != nil {
		log.Printf("analytics: recordKnockout: %v", err)
	}
}

// QueryTopPokemon returns the top-n most-used Pokemon ordered by use count descending.
func (s *AnalyticsStore) QueryTopPokemon(limit int) ([]PokemonUsageEntry, error) {
	rows, err := s.db.Query(`
		SELECT pokemon_id, name, use_count
		FROM pokemon_usage
		ORDER BY use_count DESC, pokemon_id ASC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("query top pokemon: %w", err)
	}
	defer rows.Close()

	result := []PokemonUsageEntry{}
	for rows.Next() {
		var e PokemonUsageEntry
		if err := rows.Scan(&e.PokemonID, &e.Name, &e.UseCount); err != nil {
			return nil, fmt.Errorf("scan pokemon usage: %w", err)
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

// QueryDamageTotals returns aggregated dealt and healed damage.
func (s *AnalyticsStore) QueryDamageTotals() (DamageTotalsResponse, error) {
	var r DamageTotalsResponse
	err := s.db.QueryRow(`
		SELECT total_dealt, total_healed FROM damage_totals WHERE id = 1
	`).Scan(&r.TotalDealt, &r.TotalHealed)
	if err != nil {
		return r, fmt.Errorf("query damage totals: %w", err)
	}
	return r, nil
}

// QueryKnockouts returns the total knockout count.
func (s *AnalyticsStore) QueryKnockouts() (KnockoutTotalResponse, error) {
	var r KnockoutTotalResponse
	err := s.db.QueryRow(`
		SELECT total_knockouts FROM damage_totals WHERE id = 1
	`).Scan(&r.TotalKnockouts)
	if err != nil {
		return r, fmt.Errorf("query knockouts: %w", err)
	}
	return r, nil
}
