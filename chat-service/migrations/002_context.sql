ALTER TABLE messages ADD COLUMN IF NOT EXISTS client_message_id TEXT;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reply_to BIGINT REFERENCES messages(id) ON DELETE CASCADE;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS sources JSONB NOT NULL DEFAULT '[]'::jsonb;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS suggested_action JSONB;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS mode TEXT NOT NULL DEFAULT 'mock';
CREATE UNIQUE INDEX IF NOT EXISTS messages_client_key ON messages(chat_id,client_message_id);
CREATE UNIQUE INDEX IF NOT EXISTS messages_reply_key ON messages(reply_to);
