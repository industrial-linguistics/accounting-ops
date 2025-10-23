CREATE TABLE IF NOT EXISTS auth_session (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  state TEXT NOT NULL,
  code_verifier TEXT,
  realm_id TEXT,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  ready_at INTEGER,
  result_cipher BLOB,
  consumed INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_auth_session_exp ON auth_session(expires_at);

CREATE TABLE IF NOT EXISTS rate_limit (
  key TEXT PRIMARY KEY,
  window_start INTEGER NOT NULL,
  count INTEGER NOT NULL
);
