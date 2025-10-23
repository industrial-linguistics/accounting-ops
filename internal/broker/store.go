package broker

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

var (
	//go:embed sql/schema.sql
	schemaSQL string
)

// Session represents a short-lived OAuth flow.
type Session struct {
	ID           string
	Provider     string
	State        string
	CodeVerifier sql.NullString
	RealmID      sql.NullString
	CreatedAt    time.Time
	ExpiresAt    time.Time
	ReadyAt      sql.NullTime
	Result       []byte
}

// Store wraps SQLite persistence for session management.
type Store struct {
	db *sql.DB
}

// OpenStore opens (and initialises) the session store database.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite", fmt.Sprintf("file:%s?_busy_timeout=5000&_pragma=journal_mode(WAL)", path))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close releases the underlying database handle.
func (s *Store) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// InsertSession creates a new session row.
func (s *Store) InsertSession(ctx context.Context, sess Session) error {
	_, err := s.db.ExecContext(ctx, `
        INSERT INTO auth_session(id, provider, state, code_verifier, realm_id, created_at, expires_at)
        VALUES(?, ?, ?, ?, ?, ?, ?)
    `, sess.ID, sess.Provider, sess.State, nullableString(sess.CodeVerifier), nullableString(sess.RealmID), sess.CreatedAt.Unix(), sess.ExpiresAt.Unix())
	if err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

// MarkReady stores the session result payload and marks the session ready.
func (s *Store) MarkReady(ctx context.Context, sessionID string, payload []byte, realmID *string) error {
	var realm sql.NullString
	if realmID != nil {
		realm = sql.NullString{String: *realmID, Valid: true}
	}
	res, err := s.db.ExecContext(ctx, `
        UPDATE auth_session
           SET ready_at = ?, result_cipher = ?, realm_id = COALESCE(?, realm_id)
         WHERE id = ?
    `, time.Now().Unix(), payload, nullableString(realm), sessionID)
	if err != nil {
		return fmt.Errorf("mark ready: %w", err)
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// LookupByState finds a pending session by provider and state value.
func (s *Store) LookupByState(ctx context.Context, provider, state string) (*Session, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT id, provider, state, code_verifier, realm_id, created_at, expires_at, ready_at, result_cipher
          FROM auth_session
         WHERE provider = ? AND state = ?
         ORDER BY created_at DESC
         LIMIT 1
    `, provider, state)
	return scanSession(row)
}

// LoadForPoll retrieves the session for polling.
func (s *Store) LoadForPoll(ctx context.Context, sessionID string) (*Session, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT id, provider, state, code_verifier, realm_id, created_at, expires_at, ready_at, result_cipher
          FROM auth_session
         WHERE id = ?
    `, sessionID)
	return scanSession(row)
}

// Delete removes a session entirely.
func (s *Store) Delete(ctx context.Context, sessionID string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM auth_session WHERE id = ?`, sessionID)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func scanSession(row *sql.Row) (*Session, error) {
	var sess Session
	var created, expires sql.NullInt64
	var ready sql.NullInt64
	err := row.Scan(&sess.ID, &sess.Provider, &sess.State, &sess.CodeVerifier, &sess.RealmID, &created, &expires, &ready, &sess.Result)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("scan session: %w", err)
	}
	if created.Valid {
		sess.CreatedAt = time.Unix(created.Int64, 0)
	}
	if expires.Valid {
		sess.ExpiresAt = time.Unix(expires.Int64, 0)
	}
	if ready.Valid {
		sess.ReadyAt = sql.NullTime{Time: time.Unix(ready.Int64, 0), Valid: true}
	}
	return &sess, nil
}

func nullableString(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}
