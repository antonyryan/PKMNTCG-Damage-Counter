package analytics

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"pkmntcg/backend/internal/session"

	_ "modernc.org/sqlite"
)

const healGracePeriod = 2 * time.Second

type cmdTag uint8

const (
	tagRecordPokemonUsed cmdTag = iota
	tagRecordDamage
	tagRecordKnockout
	tagQueryTopPokemon
	tagQueryDamageTotals
	tagQueryKnockouts
)

type cmd struct {
	tag cmdTag

	pokemonID int
	name      string

	sessionID string
	action    session.ActionRequest
	amount    int
	now       time.Time

	limit int
	resp  chan result
}

type result struct {
	topPokemon   []PokemonUsageEntry
	damageTotals DamageTotalsResponse
	knockouts    KnockoutTotalResponse
	err          error
}

type Store struct {
	cmdCh chan cmd
	done  chan struct{}
}

func MustNew(dbPath string) *Store {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("analytics: open db %q: %v", dbPath, err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schemaSQL); err != nil {
		log.Fatalf("analytics: init schema: %v", err)
	}

	s := &Store{cmdCh: make(chan cmd, 256), done: make(chan struct{})}
	go s.run(db)
	return s
}

func (s *Store) Shutdown() {
	close(s.cmdCh)
	<-s.done
}

func (s *Store) run(db *sql.DB) {
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("analytics: db close: %v", err)
		}
		close(s.done)
	}()

	for c := range s.cmdCh {
		switch c.tag {
		case tagRecordPokemonUsed:
			sqlRecordPokemonUsed(db, c.pokemonID, c.name)
		case tagRecordDamage:
			sqlRecordDamage(db, c.sessionID, c.action, c.amount, c.now)
		case tagRecordKnockout:
			sqlRecordKnockout(db)
		case tagQueryTopPokemon:
			entries, err := sqlQueryTopPokemon(db, c.limit)
			c.resp <- result{topPokemon: entries, err: err}
		case tagQueryDamageTotals:
			r, err := sqlQueryDamageTotals(db)
			c.resp <- result{damageTotals: r, err: err}
		case tagQueryKnockouts:
			r, err := sqlQueryKnockouts(db)
			c.resp <- result{knockouts: r, err: err}
		}
	}
}

func (s *Store) Record(sessionID string, req session.ActionRequest, resolver NameResolver, now time.Time) {
	switch req.Type {
	case session.ActionSetPokemon, session.ActionEvolve:
		if req.PokemonID == nil {
			return
		}
		name := ""
		if resolvedName, ok := resolver.NameByID(*req.PokemonID); ok {
			name = resolvedName
		}
		s.cmdCh <- cmd{tag: tagRecordPokemonUsed, pokemonID: *req.PokemonID, name: name}
	case session.ActionAdjust:
		if req.Amount != nil {
			s.cmdCh <- cmd{tag: tagRecordDamage, sessionID: sessionID, action: req, amount: *req.Amount, now: now}
		}
	case session.ActionKnockout:
		s.cmdCh <- cmd{tag: tagRecordKnockout}
	}
}

func (s *Store) QueryTopPokemon(limit int) ([]PokemonUsageEntry, error) {
	resp := make(chan result, 1)
	s.cmdCh <- cmd{tag: tagQueryTopPokemon, limit: limit, resp: resp}
	r := <-resp
	return r.topPokemon, r.err
}

func (s *Store) QueryDamageTotals() (DamageTotalsResponse, error) {
	resp := make(chan result, 1)
	s.cmdCh <- cmd{tag: tagQueryDamageTotals, resp: resp}
	r := <-resp
	return r.damageTotals, r.err
}

func (s *Store) QueryKnockouts() (KnockoutTotalResponse, error) {
	resp := make(chan result, 1)
	s.cmdCh <- cmd{tag: tagQueryKnockouts, resp: resp}
	r := <-resp
	return r.knockouts, r.err
}

func slotBenchKey(req session.ActionRequest) int {
	if req.BenchIndex == nil {
		return -1
	}
	return *req.BenchIndex
}

func sqlRecordPokemonUsed(db *sql.DB, pokemonID int, name string) {
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

func sqlRecordDamage(db *sql.DB, sessionID string, req session.ActionRequest, amount int, now time.Time) {
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

func sqlRecordKnockout(db *sql.DB) {
	if _, err := db.Exec(`
		UPDATE damage_totals SET total_knockouts = total_knockouts + 1 WHERE id = 1
	`); err != nil {
		log.Printf("analytics: recordKnockout: %v", err)
	}
}

func sqlQueryTopPokemon(db *sql.DB, limit int) ([]PokemonUsageEntry, error) {
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

func sqlQueryDamageTotals(db *sql.DB) (DamageTotalsResponse, error) {
	var r DamageTotalsResponse
	err := db.QueryRow(`SELECT total_dealt, total_healed FROM damage_totals WHERE id = 1`).Scan(&r.TotalDealt, &r.TotalHealed)
	if err != nil {
		return r, fmt.Errorf("query damage totals: %w", err)
	}
	return r, nil
}

func sqlQueryKnockouts(db *sql.DB) (KnockoutTotalResponse, error) {
	var r KnockoutTotalResponse
	err := db.QueryRow(`SELECT total_knockouts FROM damage_totals WHERE id = 1`).Scan(&r.TotalKnockouts)
	if err != nil {
		return r, fmt.Errorf("query knockouts: %w", err)
	}
	return r, nil
}
