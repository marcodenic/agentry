package memstore

import (
	"context"
	"database/sql"
	"time"

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
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS kv_meta (bucket TEXT, key TEXT, updated INTEGER, PRIMARY KEY(bucket,key))`); err != nil {
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Close() error { return s.db.Close() }

func (s *SQLite) Set(ctx context.Context, bucket, key string, val []byte) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO kv(bucket,key,value) VALUES(?,?,?)`, bucket, key, val); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO kv_meta(bucket,key,updated) VALUES(?,?,?)`, bucket, key, time.Now().Unix()); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
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

// Cleanup removes entries older than the provided TTL from the given bucket.
func (s *SQLite) Cleanup(ctx context.Context, bucket string, ttl time.Duration) error {
	before := time.Now().Add(-ttl).Unix()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM kv WHERE bucket=? AND key IN (SELECT key FROM kv_meta WHERE bucket=? AND updated<?)`, bucket, bucket, before); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM kv_meta WHERE bucket=? AND updated<?`, bucket, before); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
