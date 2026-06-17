package rssfilter

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

type ItemStore struct {
	db    *sql.DB
	limit int
}

func OpenItemStore(path string, limit int) (*ItemStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	store := &ItemStore{db: db, limit: limit}
	if err := store.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *ItemStore) Close() error {
	return s.db.Close()
}

func (s *ItemStore) init(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS fetched_items (
	key TEXT PRIMARY KEY,
	title TEXT NOT NULL DEFAULT '',
	guid TEXT NOT NULL DEFAULT '',
	fetched_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS fetched_items_fetched_at_idx ON fetched_items(fetched_at);
`)
	return err
}

func (s *ItemStore) Seen(ctx context.Context, key string) (bool, error) {
	var found int
	err := s.db.QueryRowContext(ctx, `SELECT 1 FROM fetched_items WHERE key = ? LIMIT 1`, key).
		Scan(&found)
	if err == nil {
		return true, nil
	}
	if err == sql.ErrNoRows {
		return false, nil
	}
	return false, err
}

func (s *ItemStore) RecordFetched(ctx context.Context, items []Item) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	stmt, err := tx.PrepareContext(ctx, `
INSERT INTO fetched_items (key, title, guid, fetched_at)
VALUES (?, ?, ?, ?)
ON CONFLICT(key) DO UPDATE SET
	title = excluded.title,
	guid = excluded.guid,
	fetched_at = excluded.fetched_at
`)
	if err != nil {
		return err
	}
	fetchedAt := time.Now().UnixNano()
	for index, item := range items {
		key := ItemKey(item)
		if _, err := stmt.ExecContext(ctx, key, item.Title, item.GUID, fetchedAt+int64(index)); err != nil {
			return fmt.Errorf("record fetched item: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
DELETE FROM fetched_items
WHERE key NOT IN (
	SELECT key FROM fetched_items
	ORDER BY fetched_at DESC
	LIMIT ?
)
	`, s.limit); err != nil {
		return err
	}
	if err := stmt.Close(); err != nil {
		return err
	}

	return tx.Commit()
}
