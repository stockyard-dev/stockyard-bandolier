package store

import (
	"database/sql"
	"fmt"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"time"
)

type DB struct{ db *sql.DB }
type EnvVar struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Environment string `json:"environment"`
	Project     string `json:"project"`
	Sensitive   int    `json:"sensitive"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

func Open(d string) (*DB, error) {
	if err := os.MkdirAll(d, 0755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.Join(d, "bandolier.db")+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.Exec(`CREATE TABLE IF NOT EXISTS env_vars(id TEXT PRIMARY KEY,key TEXT NOT NULL,value TEXT DEFAULT '',environment TEXT DEFAULT 'development',project TEXT DEFAULT '',sensitive INTEGER DEFAULT 0,description TEXT DEFAULT '',created_at TEXT DEFAULT(datetime('now')))`)
	db.Exec(`CREATE TABLE IF NOT EXISTS extras(
	resource TEXT NOT NULL,
	record_id TEXT NOT NULL,
	data TEXT NOT NULL DEFAULT '{}',
	PRIMARY KEY(resource, record_id)
)`)
	return &DB{db: db}, nil
}
func (d *DB) Close() error { return d.db.Close() }
func genID() string        { return fmt.Sprintf("%d", time.Now().UnixNano()) }
func now() string          { return time.Now().UTC().Format(time.RFC3339) }
func (d *DB) Create(e *EnvVar) error {
	e.ID = genID()
	e.CreatedAt = now()
	_, err := d.db.Exec(`INSERT INTO env_vars(id,key,value,environment,project,sensitive,description,created_at)VALUES(?,?,?,?,?,?,?,?)`, e.ID, e.Key, e.Value, e.Environment, e.Project, e.Sensitive, e.Description, e.CreatedAt)
	return err
}
func (d *DB) Get(id string) *EnvVar {
	var e EnvVar
	if d.db.QueryRow(`SELECT id,key,value,environment,project,sensitive,description,created_at FROM env_vars WHERE id=?`, id).Scan(&e.ID, &e.Key, &e.Value, &e.Environment, &e.Project, &e.Sensitive, &e.Description, &e.CreatedAt) != nil {
		return nil
	}
	return &e
}
func (d *DB) List() []EnvVar {
	rows, _ := d.db.Query(`SELECT id,key,value,environment,project,sensitive,description,created_at FROM env_vars ORDER BY created_at DESC`)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []EnvVar
	for rows.Next() {
		var e EnvVar
		rows.Scan(&e.ID, &e.Key, &e.Value, &e.Environment, &e.Project, &e.Sensitive, &e.Description, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}
func (d *DB) Update(e *EnvVar) error {
	_, err := d.db.Exec(`UPDATE env_vars SET key=?,value=?,environment=?,project=?,sensitive=?,description=? WHERE id=?`, e.Key, e.Value, e.Environment, e.Project, e.Sensitive, e.Description, e.ID)
	return err
}
func (d *DB) Delete(id string) error {
	_, err := d.db.Exec(`DELETE FROM env_vars WHERE id=?`, id)
	return err
}
func (d *DB) Count() int {
	var n int
	d.db.QueryRow(`SELECT COUNT(*) FROM env_vars`).Scan(&n)
	return n
}

func (d *DB) Search(q string, filters map[string]string) []EnvVar {
	where := "1=1"
	args := []any{}
	if q != "" {
		where += " AND (key LIKE ? OR description LIKE ?)"
		args = append(args, "%"+q+"%")
		args = append(args, "%"+q+"%")
	}
	if v, ok := filters["environment"]; ok && v != "" {
		where += " AND environment=?"
		args = append(args, v)
	}
	rows, _ := d.db.Query(`SELECT id,key,value,environment,project,sensitive,description,created_at FROM env_vars WHERE `+where+` ORDER BY created_at DESC`, args...)
	if rows == nil {
		return nil
	}
	defer rows.Close()
	var o []EnvVar
	for rows.Next() {
		var e EnvVar
		rows.Scan(&e.ID, &e.Key, &e.Value, &e.Environment, &e.Project, &e.Sensitive, &e.Description, &e.CreatedAt)
		o = append(o, e)
	}
	return o
}

func (d *DB) Stats() map[string]any {
	m := map[string]any{"total": d.Count()}
	return m
}

// ─── Extras: generic key-value storage for personalization custom fields ───

func (d *DB) GetExtras(resource, recordID string) string {
	var data string
	err := d.db.QueryRow(
		`SELECT data FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	).Scan(&data)
	if err != nil || data == "" {
		return "{}"
	}
	return data
}

func (d *DB) SetExtras(resource, recordID, data string) error {
	if data == "" {
		data = "{}"
	}
	_, err := d.db.Exec(
		`INSERT INTO extras(resource, record_id, data) VALUES(?, ?, ?)
		 ON CONFLICT(resource, record_id) DO UPDATE SET data=excluded.data`,
		resource, recordID, data,
	)
	return err
}

func (d *DB) DeleteExtras(resource, recordID string) error {
	_, err := d.db.Exec(
		`DELETE FROM extras WHERE resource=? AND record_id=?`,
		resource, recordID,
	)
	return err
}

func (d *DB) AllExtras(resource string) map[string]string {
	out := make(map[string]string)
	rows, _ := d.db.Query(
		`SELECT record_id, data FROM extras WHERE resource=?`,
		resource,
	)
	if rows == nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var id, data string
		rows.Scan(&id, &data)
		out[id] = data
	}
	return out
}
