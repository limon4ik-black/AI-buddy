CREATE TABLE login_confirmations (
 token_hash text PRIMARY KEY, user_id text NOT NULL REFERENCES users(id),
 expires_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX login_confirmation_expiry ON login_confirmations(expires_at);
