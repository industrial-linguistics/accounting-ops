CREATE TABLE IF NOT EXISTS auth_session (
  id TEXT PRIMARY KEY,
  provider TEXT NOT NULL,
  state TEXT NOT NULL,
  code_verifier TEXT,
  realm_id TEXT,
  created_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  ready_at INTEGER,
  result_cipher BLOB
);

CREATE INDEX IF NOT EXISTS idx_auth_session_exp ON auth_session(expires_at);
