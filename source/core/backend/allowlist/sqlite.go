package allowlist

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

var (
	ErrNotFound = errors.New("allowlist: program not found")
	ErrDisabled = errors.New("allowlist: program disabled")
	ErrExists   = errors.New("allowlist: program already exists")
)

type Program struct {
	Name    string
	Command string
	Args    []string
	Enabled bool
}

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS programs (
	name    TEXT PRIMARY KEY,
	command TEXT NOT NULL,
	args    TEXT NOT NULL DEFAULT '[]',
	enabled INTEGER NOT NULL DEFAULT 1
);`

func Open(path string) (*Store, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"

	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("allowlist: open %s: %w", path, err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("allowlist: connect %s: %w", path, err)
	}

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("allowlist: create schema: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Get liefert einen aktiven Eintrag. Fehlt er oder ist er deaktiviert,
// kommt ErrNotFound bzw. ErrDisabled zurück.
func (s *Store) Get(name string) (Program, error) {
	var (
		p       Program
		argsRaw string
		enabled int
	)

	err := s.db.QueryRow(
		`SELECT name, command, args, enabled FROM programs WHERE name = ?`, name,
	).Scan(&p.Name, &p.Command, &argsRaw, &enabled)
	if errors.Is(err, sql.ErrNoRows) {
		return Program{}, ErrNotFound
	}
	if err != nil {
		return Program{}, fmt.Errorf("allowlist: get %q: %w", name, err)
	}

	if err := json.Unmarshal([]byte(argsRaw), &p.Args); err != nil {
		return Program{}, fmt.Errorf("allowlist: decode args of %q: %w", name, err)
	}
	p.Enabled = enabled == 1

	if !p.Enabled {
		return p, ErrDisabled
	}
	return p, nil
}

// List liefert alle Einträge, auch deaktivierte, sortiert nach Name.
func (s *Store) List() ([]Program, error) {
	rows, err := s.db.Query(`SELECT name, command, args, enabled FROM programs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("allowlist: list: %w", err)
	}
	defer rows.Close()

	var programs []Program
	for rows.Next() {
		var (
			p       Program
			argsRaw string
			enabled int
		)
		if err := rows.Scan(&p.Name, &p.Command, &argsRaw, &enabled); err != nil {
			return nil, fmt.Errorf("allowlist: scan: %w", err)
		}
		if err := json.Unmarshal([]byte(argsRaw), &p.Args); err != nil {
			return nil, fmt.Errorf("allowlist: decode args of %q: %w", p.Name, err)
		}
		p.Enabled = enabled == 1
		programs = append(programs, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("allowlist: list rows: %w", err)
	}

	return programs, nil
}

// Add legt einen neuen Eintrag an. Existiert der Name bereits, kommt ErrExists.
func (s *Store) Add(p Program) error {
	if p.Name == "" || p.Command == "" {
		return errors.New("allowlist: name and command must not be empty")
	}

	args := p.Args
	if args == nil {
		args = []string{}
	}
	argsRaw, err := json.Marshal(args)
	if err != nil {
		return fmt.Errorf("allowlist: encode args: %w", err)
	}

	enabled := 0
	if p.Enabled {
		enabled = 1
	}

	res, err := s.db.Exec(
		`INSERT INTO programs (name, command, args, enabled) VALUES (?, ?, ?, ?)
		 ON CONFLICT(name) DO NOTHING`,
		p.Name, p.Command, string(argsRaw), enabled,
	)
	if err != nil {
		return fmt.Errorf("allowlist: add %q: %w", p.Name, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("allowlist: add %q: %w", p.Name, err)
	}
	if n == 0 {
		return ErrExists
	}
	return nil
}

func (s *Store) Remove(name string) error {
	res, err := s.db.Exec(`DELETE FROM programs WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("allowlist: remove %q: %w", name, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("allowlist: remove %q: %w", name, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) SetEnabled(name string, enabled bool) error {
	value := 0
	if enabled {
		value = 1
	}

	res, err := s.db.Exec(`UPDATE programs SET enabled = ? WHERE name = ?`, value, name)
	if err != nil {
		return fmt.Errorf("allowlist: set enabled %q: %w", name, err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("allowlist: set enabled %q: %w", name, err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
