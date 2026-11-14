CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS documents (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    chunker TEXT NOT NULL,
    tokenizer TEXT NOT NULL,
    max_tokens INTEGER NOT NULL,
    embedding_model TEXT NOT NULL DEFAULT 'Qwen/Qwen3-Embedding-0.6B',
    embedding_version TEXT,
    indexed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS chunks (
    id TEXT PRIMARY KEY,
    document_id TEXT NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    headings JSONB NOT NULL DEFAULT '[]'::jsonb,
    text TEXT NOT NULL,
    contextualized_text TEXT NOT NULL,
    contextualized_tokens INTEGER NOT NULL,
    embedding VECTOR(1024) NOT NULL,

    UNIQUE (document_id, chunk_index)
);

CREATE INDEX IF NOT EXISTS idx_chunks_fts
ON chunks
USING GIN (to_tsvector('simple', contextualized_text));

CREATE INDEX IF NOT EXISTS idx_chunks_document_id
ON chunks (document_id);
