package analytics

const schemaSQL = `
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
