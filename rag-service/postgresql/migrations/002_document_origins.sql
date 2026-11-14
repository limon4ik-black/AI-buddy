ALTER TABLE documents ADD COLUMN origin TEXT NOT NULL DEFAULT 'note_service'
    CHECK (origin IN ('note_service', 'bundled_json', 'legacy_json'));
ALTER TABLE documents ADD COLUMN dataset_name TEXT;

-- The old JSON importer never set embedding_version; Note's indexer always did.
-- Keep unknown legacy data until explicitly adopted by the bundled importer.
UPDATE documents SET origin = 'legacy_json' WHERE embedding_version IS NULL;

CREATE INDEX idx_documents_origin_dataset ON documents(origin, dataset_name);

CREATE TABLE knowledge_imports (
    dataset_name TEXT PRIMARY KEY,
    version TEXT NOT NULL,
    manifest_sha256 TEXT NOT NULL,
    document_count INTEGER NOT NULL,
    chunk_count INTEGER NOT NULL,
    imported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
