CREATE TABLE IF NOT EXISTS chats (
 id UUID PRIMARY KEY,
 user_id UUID NOT NULL,
 title VARCHAR(255) NOT NULL DEFAULT 'Новый чат',
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS messages (
 id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 chat_id UUID NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
 role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant')),
 content TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_chats_user_updated ON chats(user_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages(chat_id, id);
