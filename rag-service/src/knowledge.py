"""Portable precomputed knowledge dataset: validate, migrate, and import atomically."""
import argparse
import hashlib
import json
import os
from pathlib import Path
from typing import Literal

import psycopg
from pgvector.psycopg import register_vector
from pydantic import BaseModel, ConfigDict, Field, StrictInt, StrictStr, field_validator

from .config import EMBEDDING_MODEL, ROOT
from .migrations import INDEX_LOCK, migrate

DEFAULT_DIRECTORY = ROOT / 'data' / 'knowledge' / 'v1'


class Chunk(BaseModel):
    model_config = ConfigDict(extra='forbid')
    chunk_id: StrictStr = Field(min_length=1)
    chunk_index: StrictInt = Field(ge=0)
    headings: list[StrictStr]
    text: StrictStr
    contextualized_text: StrictStr
    contextualized_tokens: StrictInt = Field(ge=0)
    embedding: list[float] = Field(min_length=1024, max_length=1024)

    @field_validator('embedding', mode='before')
    @classmethod
    def valid_vector(cls, value):
        import math
        if (not isinstance(value, list) or not value or
                any(type(x) not in (int, float) or not math.isfinite(x) for x in value) or
                not any(x != 0 for x in value)):
            raise ValueError('Invalid embedding')
        return value


class Document(BaseModel):
    model_config = ConfigDict(extra='forbid')
    document_id: StrictStr = Field(min_length=1)
    source: StrictStr
    chunker: StrictStr
    tokenizer: StrictStr
    max_tokens: StrictInt = Field(gt=0)
    embedding_model: Literal['Qwen/Qwen3-Embedding-0.6B']
    embedding_dimensions: Literal[1024]
    chunks: list[Chunk] = Field(min_length=1)


class FileEntry(BaseModel):
    model_config = ConfigDict(extra='forbid')
    file: StrictStr
    sha256: str = Field(pattern=r'^[0-9a-f]{64}$')
    document_id: StrictStr
    chunks: StrictInt = Field(gt=0)

    @field_validator('file')
    @classmethod
    def simple_filename(cls, name):
        if Path(name).name != name or '\\' in name or not name.endswith('.json') or name == 'manifest.json':
            raise ValueError('Expected a JSON filename without directories')
        return name


class Manifest(BaseModel):
    model_config = ConfigDict(extra='forbid')
    schema_version: Literal[1]
    dataset_name: StrictStr = Field(min_length=1)
    version: StrictStr = Field(min_length=1)
    embedding_model: Literal['Qwen/Qwen3-Embedding-0.6B']
    embedding_dimensions: Literal[1024]
    visibility: Literal['company']
    files: list[FileEntry] = Field(min_length=1)


def knowledge_directory():
    return Path(os.environ.get('RAG_KNOWLEDGE_DIR', str(DEFAULT_DIRECTORY)))


def load_dataset(directory):
    directory = Path(directory)
    manifest_bytes = (directory / 'manifest.json').read_bytes()
    manifest = Manifest.model_validate_json(manifest_bytes)
    expected = [entry.file for entry in manifest.files]
    actual = {path.name for path in directory.glob('*.json') if path.name != 'manifest.json'}
    if len(set(expected)) != len(expected) or set(expected) != actual:
        raise ValueError('Manifest must list every dataset JSON file exactly once')
    documents, ids, chunks = [], set(), set()
    for entry in manifest.files:
        content = (directory / entry.file).read_bytes()
        if hashlib.sha256(content).hexdigest() != entry.sha256:
            raise ValueError(f'Checksum mismatch: {entry.file}')
        document = Document.model_validate_json(content)
        if document.document_id != entry.document_id or len(document.chunks) != entry.chunks:
            raise ValueError(f'Manifest metadata mismatch: {entry.file}')
        if document.document_id in ids:
            raise ValueError('Duplicate document ID')
        ids.add(document.document_id)
        indexes = set()
        for chunk in document.chunks:
            if chunk.chunk_id in chunks or chunk.chunk_index in indexes:
                raise ValueError('Duplicate chunk ID or index')
            chunks.add(chunk.chunk_id)
            indexes.add(chunk.chunk_index)
        documents.append((entry, document))
    return manifest, hashlib.sha256(manifest_bytes).hexdigest(), documents


def import_dataset(conn, directory):
    # Validate the entire snapshot before any writes, including migration writes.
    manifest, checksum, documents = load_dataset(directory)
    migrate(conn)
    register_vector(conn)
    changed = 0
    with conn.transaction():
        conn.execute('SELECT pg_advisory_xact_lock(%s)', (INDEX_LOCK,))
        previous = {row[0]: row[1:] for row in conn.execute(
            'SELECT id, origin, dataset_name, embedding_version FROM documents').fetchall()}
        for entry, doc in documents:
            old = previous.get(doc.document_id)
            if old and old[0] != 'legacy_json' and old[:2] != ('bundled_json', manifest.dataset_name):
                raise ValueError(f'Document ID belongs to another source: {doc.document_id}')
            if old == ('bundled_json', manifest.dataset_name, entry.sha256):
                continue
            # Replace a changed document and all its chunks together (no stale chunks).
            conn.execute('DELETE FROM documents WHERE id=%s', (doc.document_id,))
            conn.execute('''INSERT INTO documents
                (id, source, chunker, tokenizer, max_tokens, embedding_model,
                 embedding_version, origin, dataset_name)
                VALUES (%s,%s,%s,%s,%s,%s,%s,'bundled_json',%s)''',
                (doc.document_id, doc.source, doc.chunker, doc.tokenizer, doc.max_tokens,
                 manifest.embedding_model, entry.sha256, manifest.dataset_name))
            for chunk in doc.chunks:
                conn.execute('''INSERT INTO chunks
                    (id,document_id,chunk_index,headings,text,contextualized_text,contextualized_tokens,embedding)
                    VALUES (%s,%s,%s,%s::jsonb,%s,%s,%s,%s)''',
                    (chunk.chunk_id, doc.document_id, chunk.chunk_index,
                     json.dumps(chunk.headings, ensure_ascii=False), chunk.text,
                     chunk.contextualized_text, chunk.contextualized_tokens, chunk.embedding))
            changed += 1
        conn.execute('''DELETE FROM documents WHERE origin='bundled_json' AND dataset_name=%s
                        AND NOT (id = ANY(%s::text[]))''',
                     (manifest.dataset_name, [doc.document_id for _, doc in documents]))
        conn.execute('''INSERT INTO knowledge_imports
            (dataset_name,version,manifest_sha256,document_count,chunk_count)
            VALUES (%s,%s,%s,%s,%s) ON CONFLICT (dataset_name) DO UPDATE SET
            version=EXCLUDED.version, manifest_sha256=EXCLUDED.manifest_sha256,
            document_count=EXCLUDED.document_count,chunk_count=EXCLUDED.chunk_count,imported_at=NOW()
            WHERE knowledge_imports.manifest_sha256 <> EXCLUDED.manifest_sha256''',
            (manifest.dataset_name, manifest.version, checksum, len(documents), sum(e.chunks for e, _ in documents)))
    return {'dataset': manifest.dataset_name, 'version': manifest.version, 'documents': len(documents),
            'chunks': sum(e.chunks for e, _ in documents), 'changed_documents': changed}


def bootstrap(db_kwargs, directory=None):
    with psycopg.connect(**db_kwargs, autocommit=True) as conn:
        return import_dataset(conn, directory or knowledge_directory())


def write_manifest(directory, version, dataset_name='company-knowledge'):
    directory = Path(directory)
    files = []
    for path in sorted(directory.glob('*.json')):
        if path.name == 'manifest.json':
            continue
        content = path.read_bytes()
        doc = Document.model_validate_json(content)
        files.append({'file': path.name, 'sha256': hashlib.sha256(content).hexdigest(),
                      'document_id': doc.document_id, 'chunks': len(doc.chunks)})
    manifest = Manifest(schema_version=1, dataset_name=dataset_name, version=version,
                        embedding_model=EMBEDDING_MODEL, embedding_dimensions=1024,
                        visibility='company', files=files)
    # Detect duplicate IDs across files before replacing the manifest.
    if len({e.document_id for e in manifest.files}) != len(manifest.files):
        raise ValueError('Duplicate document ID')
    (directory / 'manifest.json').write_text(manifest.model_dump_json(indent=2) + '\n', encoding='utf-8')
    return load_dataset(directory)[0]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('command', choices=['validate', 'import', 'manifest'])
    parser.add_argument('directory', nargs='?', type=Path, default=knowledge_directory())
    parser.add_argument('--version', default='v1')
    parser.add_argument('--dataset-name', default='company-knowledge')
    args = parser.parse_args()
    if args.command == 'manifest':
        write_manifest(args.directory, args.version, args.dataset_name)
    if args.command != 'import':
        manifest, _, documents = load_dataset(args.directory)
        print(f'{manifest.dataset_name}/{manifest.version}: {len(documents)} documents, '
              f'{sum(e.chunks for e, _ in documents)} chunks; validated')
        return
    # Import needs only DB credentials, not a Groq key or a loaded embedding model.
    from dotenv import load_dotenv
    load_dotenv(ROOT / '.env')
    kwargs = dict(host=os.environ['POSTGRES_HOST'], port=int(os.getenv('POSTGRES_PORT', '5432')),
                  user=os.environ['POSTGRES_USER'], password=os.environ['POSTGRES_PASSWORD'],
                  dbname=os.environ['POSTGRES_DB'], connect_timeout=10)
    print(json.dumps(bootstrap(kwargs, args.directory), ensure_ascii=False))


if __name__ == '__main__':
    main()
