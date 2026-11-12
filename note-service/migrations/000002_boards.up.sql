CREATE SCHEMA IF NOT EXISTS board;

CREATE TABLE board.boards (
    id         text PRIMARY KEY,
    user_id    text NOT NULL,
    title      text NOT NULL DEFAULT '' CHECK (length(title) <= 200),
    scene      jsonb NOT NULL CHECK (jsonb_typeof(scene) = 'object' AND octet_length(scene::text) <= 8388608),
    version    integer NOT NULL DEFAULT 1 CHECK (version > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX boards_owner ON board.boards(user_id, updated_at DESC, id);
