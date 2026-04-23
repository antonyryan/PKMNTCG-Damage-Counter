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

// ---------------------------------------------------------------------------
// Worker / channel layer
// ---------------------------------------------------------------------------

// cmdTag identifies which analytics operation a command represents.
type cmdTag uint8

const (
	tagRecordPokemonUsed cmdTag = iota
	tagRecordDamage
	tagRecordKnockout
	tagQueryTopPokemon
	tagQueryDamageTotals
	tagQueryKnockouts
)

// analyticsCmd is the single message type sent through the worker channel.
// Write commands leave resp nil; the worker logs errors internally.
// Read commands set resp to a buffered channel so the caller can block on it.
type analyticsCmd struct {
	tag cmdTag

	// tagRecordPokemonUsed
	pokemonID int
	name      string

	// tagRecordDamage
	sessionID string
	req       SessionActionRequest
	amount    int
	now       time.Time

	// tagQueryTopPokemon
	limit int

	// non-nil only for read commands
	resp chan analyticsResult
}

// analyticsResult carries the response for a read command.
type analyticsResult struct {
	topPokemon   []PokemonUsageEntry
	damageTotals DamageTotalsResponse
	knockouts    KnockoutTotalResponse
	err          error
}

// AnalyticsStore serialises all SQLite access through a single worker goroutine,
// eliminating SQLITE_BUSY races without relying on SetMaxOpenConns(1) alone.
type AnalyticsStore struct {
	cmdCh chan analyticsCmd
	done  chan struct{}
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

// mustNewAnalyticsStore opens (or creates) the SQLite database, initialises the
// schema, and starts the background worker goroutine.  It terminates the process
// on any unrecoverable error.
func mustNewAnalyticsStore(dbPath string) *AnalyticsStore {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("analytics: open db %q: %v", dbPath, err)
	}
	// The worker is the only goroutine that touches the DB, so one connection
	// is both sufficient and correct.
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(analyticsSchema); err != nil {
		log.Fatalf("analytics: init schema: %v", err)
	}

	s := &AnalyticsStore{
		// Buffer of 256 lets Record() callers return immediately under normal load.
		cmdCh: make(chan analyticsCmd, 256),
		done:  make(chan struct{}),
	}
	go s.run(db)
	return s
}

// Shutdown drains the command queue, stops the worker, and closes the database.
// It must be called exactly once during application shutdown.
func (s *AnalyticsStore) Shutdown() {
	close(s.cmdCh)
	<-s.done
}

// run is the worker goroutine.  It is the only place where db is accessed.
func (s *AnalyticsStore) run(db *sql.DB) {
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("analytics: db close: %v", err)
		}
		close(s.done)
	}()

	for cmd := range s.cmdCh {
		switch cmd.tag {
		case tagRecordPokemonUsed:
			workerRecordPokemonUsed(db, cmd.pokemonID, cmd.name)
		case tagRecordDamage:
			workerRecordDamage(db, cmd.sessionID, cmd.req, cmd.amount, cmd.now)
		case tagRecordKnockout:
			workerRecordKnockout(db)
		case tagQueryTopPokemon:
			entries, err := workerQueryTopPokemon(db, cmd.limit)
			cmd.resp <- analyticsResult{topPokemon: entries, err: err}
		case tagQueryDamageTotals:
			r, err := workerQueryDamageTotals(db)
			cmd.resp <- analyticsResult{damageTotals: r, err: err}
		case tagQueryKnockouts:
			r, err := workerQueryKnockouts(db)
			cmd.resp <- analyticsResult{knockouts: r, err: err}
		}
	}
}

// ---------------------------------------------------------------------------
// Public interface – safe to call from any goroutine
// ---------------------------------------------------------------------------

// Record inspects an already-validated-and-applied action and enqueues the
// appropriate analytics write command(s).  It returns immediately; the worker
// goroutine executes the SQL and logs any errors.
// now is the wall-clock time the action was received (heal-grace-period check).
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
		s.cmdCh <- analyticsCmd{tag: tagRecordPokemonUsed, pokemonID: *req.PokemonID, name: name}

	case ActionAdjust:
		if req.Amount != nil {
			s.cmdCh <- analyticsCmd{
				tag:       tagRecordDamage,
				sessionID: sessionID,
				req:       req,
				amount:    *req.Amount,
				now:       now,
			}
		}

	case ActionKnockout:
		s.cmdCh <- analyticsCmd{tag: tagRecordKnockout}
	}
}

// QueryTopPokemon returns the top-n most-used Pokemon ordered by use count descending.
func (s *AnalyticsStore) QueryTopPokemon(limit int) ([]PokemonUsageEntry, error) {
	resp := make(chan analyticsResult, 1)
	s.cmdCh <- analyticsCmd{tag: tagQueryTopPokemon, limit: limit, resp: resp}
	r := <-resp
	return r.topPokemon, r.err
}

// QueryDamageTotals returns aggregated dealt and healed damage.
func (s *AnalyticsStore) QueryDamageTotals() (DamageTotalsResponse, error) {
	resp := make(chan analyticsResult, 1)
	s.cmdCh <- analyticsCmd{tag: tagQueryDamageTotals, resp: resp}
	r := <-resp
	return r.damageTotals, r.err
}

// QueryKnockouts returns the total knockout count.
func (s *AnalyticsStore) QueryKnockouts() (KnockoutTotalResponse, error) {
	resp := make(chan analyticsResult, 1)
	s.cmdCh <- analyticsCmd{tag: tagQueryKnockouts, resp: resp}
	r := <-resp
	return r.knockouts, r.err
}

// ---------------------------------------------------------------------------
// Worker-side SQL helpers – called only from run(), never from other goroutines
// ---------------------------------------------------------------------------

func workerRecordPokemonUsed(db *sql.DB, pokemonID int, name string) {
	_, err := db.Exec(`
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

func workerRecordDamage(db *sql.DB, sessionID string, req SessionActionRequest, amount int, now time.Time) {
	bk := slotBenchKey(req)
	side := string(req.Side)
	zone := string(req.Zone)

	if amount > 0 {
		if _, err := db.Exec(`
			UPDATE damage_totals SET total_dealt = total_dealt + ? WHERE id = 1
		`, amount); err != nil {
			log.Printf("analytics: total_dealt += %d: %v", amount, err)
			return
		}
		if _, err := db.Exec(`
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
		var lastMs int64
		err := db.QueryRow(`
			SELECT last_positive_at FROM slot_last_damage
			WHERE session_id = ? AND side = ? AND zone = ? AND bench_index = ?
		`, sessionID, side, zone, bk).Scan(&lastMs)
		if err == sql.ErrNoRows {
			return
		}
		if err != nil {
			log.Printf("analytics: slot_last_damage query: %v", err)
			return
		}
		if now.Sub(time.UnixMilli(lastMs)) <= healGracePeriod {
			return
		}
		healed := -amount
		if _, err := db.Exec(`
			UPDATE damage_totals SET total_healed = total_healed + ? WHERE id = 1
		`, healed); err != nil {
			log.Printf("analytics: total_healed += %d: %v", healed, err)
		}
	}
}

func workerRecordKnockout(db *sql.DB) {
	if _, err := db.Exec(`
		UPDATE damage_totals SET total_knockouts = total_knockouts + 1 WHERE id = 1
	`); err != nil {
		log.Printf("analytics: recordKnockout: %v", err)
	}
}

func workerQueryTopPokemon(db *sql.DB, limit int) ([]PokemonUsageEntry, error) {
	rows, err := db.Query(`
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

func workerQueryDamageTotals(db *sql.DB) (DamageTotalsResponse, error) {
	var r DamageTotalsResponse
	err := db.QueryRow(`
		SELECT total_dealt, total_healed FROM damage_totals WHERE id = 1
	`).Scan(&r.TotalDealt, &r.TotalHealed)
	if err != nil {
		return r, fmt.Errorf("query damage totals: %w", err)
	}
	return r, nil
}

func workerQueryKnockouts(db *sql.DB) (KnockoutTotalResponse, error) {
	var r KnockoutTotalResponse
	err := db.QueryRow(`
		SELECT total_knockouts FROM damage_totals WHERE id = 1
	`).Scan(&r.TotalKnockouts)
	if err != nil {
		return r, fmt.Errorf("query knockouts: %w", err)
	}
	return r, nil
}
