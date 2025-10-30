package broker

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
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
	Consumed     bool
}

// Store wraps SQLite persistence for session management.
type Store struct {
	db *sql.DB
}

// ErrRateLimited indicates a caller has exceeded the configured quota.
var ErrRateLimited = errors.New("rate limit exceeded")

// OpenStore opens (and initialises) the session store database.
func OpenStore(path string) (*Store, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_busy_timeout=5000&_pragma=journal_mode(WAL)", path))
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply schema: %w", err)
	}
	if err := ensureConsumedColumn(db); err != nil {
		db.Close()
		return nil, err
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
        INSERT INTO auth_session(id, provider, state, code_verifier, realm_id, created_at, expires_at, consumed)
        VALUES(?, ?, ?, ?, ?, ?, ?, 0)
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
           SET ready_at = ?, result_cipher = ?, realm_id = COALESCE(?, realm_id), consumed = 1
         WHERE id = ? AND consumed = 0
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
        SELECT id, provider, state, code_verifier, realm_id, created_at, expires_at, ready_at, result_cipher, consumed
          FROM auth_session
         WHERE provider = ? AND state = ? AND consumed = 0
         ORDER BY created_at DESC
         LIMIT 1
    `, provider, state)
	return scanSession(row)
}

// LoadForPoll retrieves the session for polling.
func (s *Store) LoadForPoll(ctx context.Context, sessionID string) (*Session, error) {
	row := s.db.QueryRowContext(ctx, `
        SELECT id, provider, state, code_verifier, realm_id, created_at, expires_at, ready_at, result_cipher, consumed
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
	var consumed sql.NullInt64
	err := row.Scan(&sess.ID, &sess.Provider, &sess.State, &sess.CodeVerifier, &sess.RealmID, &created, &expires, &ready, &sess.Result, &consumed)
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
	sess.Consumed = consumed.Valid && consumed.Int64 != 0
	return &sess, nil
}

func nullableString(ns sql.NullString) interface{} {
	if ns.Valid {
		return ns.String
	}
	return nil
}

func ensureConsumedColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(auth_session)`)
	if err != nil {
		return fmt.Errorf("inspect auth_session schema: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid     int
			name    string
			colType string
			notNull int
			dflt    sql.NullString
			pk      int
		)
		if scanErr := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); scanErr != nil {
			return fmt.Errorf("scan auth_session schema: %w", scanErr)
		}
		if name == "consumed" {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate auth_session schema: %w", err)
	}
	if _, err := db.Exec(`ALTER TABLE auth_session ADD COLUMN consumed INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("add consumed column: %w", err)
	}
	return nil
}

// IncrementRateLimit records a call for the provided key and enforces the configured threshold.
func (s *Store) IncrementRateLimit(ctx context.Context, key string, limit int, window time.Duration) (err error) {
	if limit <= 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rate limit tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Unix()
	windowSeconds := int64(window / time.Second)
	if windowSeconds <= 0 {
		windowSeconds = 1
	}

	var start sql.NullInt64
	var count sql.NullInt64
	queryErr := tx.QueryRowContext(ctx, `SELECT window_start, count FROM rate_limit WHERE key = ?`, key).Scan(&start, &count)
	switch {
	case errors.Is(queryErr, sql.ErrNoRows):
		if _, err = tx.ExecContext(ctx, `INSERT INTO rate_limit(key, window_start, count) VALUES(?, ?, 1)`, key, now); err != nil {
			return fmt.Errorf("insert rate limit: %w", err)
		}
	case queryErr != nil:
		return fmt.Errorf("query rate limit: %w", queryErr)
	default:
		if !start.Valid || now-start.Int64 >= windowSeconds {
			if _, err = tx.ExecContext(ctx, `UPDATE rate_limit SET window_start = ?, count = 1 WHERE key = ?`, now, key); err != nil {
				return fmt.Errorf("reset rate limit: %w", err)
			}
		} else if count.Valid && count.Int64 >= int64(limit) {
			return ErrRateLimited
		} else {
			if _, err = tx.ExecContext(ctx, `UPDATE rate_limit SET count = count + 1 WHERE key = ?`, key); err != nil {
				return fmt.Errorf("increment rate limit: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit rate limit: %w", err)
	}
	return nil
}
