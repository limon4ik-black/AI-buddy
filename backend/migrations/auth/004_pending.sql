CREATE TABLE pending_registrations (
 email text PRIMARY KEY, password_hash text NOT NULL,
 code_hash text NOT NULL, code_expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
