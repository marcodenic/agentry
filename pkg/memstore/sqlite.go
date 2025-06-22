package memstore

import (
	"context"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SQLite is the default durable implementation backed by SQLite.
type SQLite struct {
	db *sql.DB
}

func NewSQLite(path string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS kv (bucket TEXT, key TEXT, value BLOB, PRIMARY KEY(bucket,key))`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS vector (id TEXT PRIMARY KEY, text TEXT)`); err != nil {
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

func (s *SQLite) Set(ctx context.Context, bucket, key string, val []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO kv(bucket,key,value) VALUES(?,?,?)`, bucket, key, val)
	return err
}

func (s *SQLite) Get(ctx context.Context, bucket, key string) ([]byte, error) {
	row := s.db.QueryRowContext(ctx, `SELECT value FROM kv WHERE bucket=? AND key=?`, bucket, key)
	var val []byte
	if err := row.Scan(&val); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (s *SQLite) Add(ctx context.Context, id, text string) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO vector(id,text) VALUES(?,?)`, id, text)
	return err
}

func (s *SQLite) Query(ctx context.Context, text string, k int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM vector LIMIT ?`, k)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
