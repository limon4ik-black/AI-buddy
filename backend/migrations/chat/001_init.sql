CREATE TABLE chats (
 id text PRIMARY KEY, user_id text NOT NULL, title text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(), updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX chats_owner ON chats(user_id,created_at DESC,id);
CREATE TABLE messages (
 id text PRIMARY KEY, chat_id text NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
 role text NOT NULL CHECK(role IN ('user','assistant')), content text NOT NULL,
 client_message_id text, reply_to text UNIQUE REFERENCES messages(id) ON DELETE CASCADE,
 sources jsonb NOT NULL DEFAULT '[]', suggested_action jsonb,
 mode text NOT NULL DEFAULT 'mock', sequence bigserial NOT NULL UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(chat_id,client_message_id)
);
CREATE INDEX messages_chat ON messages(chat_id,sequence);
