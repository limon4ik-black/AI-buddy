CREATE SCHEMA IF NOT EXISTS note;

CREATE TABLE note.notes (
    id         text PRIMARY KEY,
    user_id    text NOT NULL,
    content    text NOT NULL CHECK (btrim(content) <> '' AND length(content) <= 20000),
    version    integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notes_owner ON note.notes(user_id, updated_at DESC, id);
