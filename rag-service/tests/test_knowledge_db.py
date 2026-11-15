"""Run with RAG_TEST_DATABASE_URL pointing to a disposable pgvector database."""
import asyncio
import os
import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import httpx
import numpy as np
import psycopg
import pytest
from pgvector.psycopg import register_vector_async
from psycopg.rows import dict_row
from psycopg_pool import AsyncConnectionPool
from pydantic import SecretStr

from src.knowledge import DEFAULT_DIRECTORY, import_dataset
from src.migrations import migrate
from src.rag import RagService
from test_knowledge import dataset, document


@pytest.fixture
def db():
    url = os.environ.get('RAG_TEST_DATABASE_URL')
    if not url:
        pytest.skip('RAG_TEST_DATABASE_URL is required for database integration tests')
    schema = 'rag_test_' + uuid.uuid4().hex
    with psycopg.connect(url, autocommit=True, connect_timeout=5) as admin:
        admin.execute('CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public')
        admin.execute(f'CREATE SCHEMA {schema}')
        kwargs = dict(conninfo=url, options=f'-c search_path={schema},public', connect_timeout=5)
        try:
            with psycopg.connect(**kwargs, autocommit=True) as conn:
                yield conn, kwargs
        finally:
            admin.execute(f'DROP SCHEMA {schema} CASCADE')


def counts(conn):
    return conn.execute('SELECT (SELECT count(*) FROM documents), (SELECT count(*) FROM chunks)').fetchone()


def test_fresh_and_repeated_real_dataset_import(db):
    conn, _ = db
    first = import_dataset(conn, DEFAULT_DIRECTORY)
    assert first['changed_documents'] == 23
    assert counts(conn) == (23, 637)
    timestamps = conn.execute('SELECT id,indexed_at FROM documents ORDER BY id').fetchall()
    imported_at = conn.execute('SELECT imported_at FROM knowledge_imports').fetchone()
    assert import_dataset(conn, DEFAULT_DIRECTORY)['changed_documents'] == 0
    assert conn.execute('SELECT id,indexed_at FROM documents ORDER BY id').fetchall() == timestamps
    assert conn.execute('SELECT imported_at FROM knowledge_imports').fetchone() == imported_at
    assert conn.execute('SELECT count(*) FROM rag_schema_migrations').fetchone() == (2,)


def test_upgrade_legacy_database_adopts_import_and_preserves_unknown_data(db, tmp_path):
    conn, _ = db
    from src.migrations import MIGRATIONS
    conn.execute((MIGRATIONS / '001_initial.sql').read_text())
    for doc_id, version in [('doc', None), ('unknown', None), ('note', 'fingerprint')]:
        conn.execute('''INSERT INTO documents(id,source,chunker,tokenizer,max_tokens,embedding_version)
                        VALUES (%s,'s','c','t',512,%s)''', (doc_id, version))
    import_dataset(conn, dataset(tmp_path))
    assert dict(conn.execute('SELECT id,origin FROM documents').fetchall()) == {
        'doc': 'bundled_json', 'unknown': 'legacy_json', 'note': 'note_service'}
    assert counts(conn) == (3, 2)
    migrate(conn)
    assert counts(conn) == (3, 2)


def test_update_removes_old_chunks_and_removed_documents_only_from_same_dataset(db, tmp_path):
    conn, _ = db
    import_dataset(conn, dataset(tmp_path, [document('a'), document('b')]))
    conn.execute("INSERT INTO documents(id,source,chunker,tokenizer,max_tokens) VALUES ('note','s','c','t',512)")
    result = import_dataset(conn, dataset(tmp_path, [document('a', 1)], version='v2'))
    assert result['changed_documents'] == 1
    assert counts(conn) == (2, 1)
    assert set(row[0] for row in conn.execute('SELECT id FROM documents')) == {'a', 'note'}
    assert conn.execute('SELECT version FROM knowledge_imports').fetchone() == ('v2',)


def test_collision_rolls_back_whole_import(db, tmp_path):
    conn, _ = db
    import_dataset(conn, dataset(tmp_path, [document('a')]))
    conn.execute("INSERT INTO documents(id,source,chunker,tokenizer,max_tokens) VALUES ('z','s','c','t',512)")
    # 'a' changes before collision on 'z': transaction must restore its old chunks.
    with pytest.raises(ValueError, match='another source'):
        import_dataset(conn, dataset(tmp_path, [document('a', 1), document('z')], version='v2'))
    assert counts(conn) == (2, 2)
    assert conn.execute('SELECT version FROM knowledge_imports').fetchone() == ('v1',)


def test_chunk_collision_rolls_back_whole_import(db, tmp_path):
    conn, _ = db
    import_dataset(conn, dataset(tmp_path, [document('a')]))
    other = document('b')
    other['chunks'][0]['chunk_id'] = 'a:0'
    # Data from another dataset cannot replace existing chunk IDs.
    path = dataset(tmp_path, [other])
    import json
    file = path / 'manifest.json'
    manifest = json.loads(file.read_text())
    manifest['dataset_name'] = 'other-dataset'
    file.write_text(json.dumps(manifest))
    with pytest.raises(psycopg.errors.UniqueViolation):
        import_dataset(conn, path)
    assert counts(conn) == (1, 2)


def test_note_sync_preserves_bundle_and_company_retrieval_keeps_note_acl(db, tmp_path, monkeypatch):
    conn, kwargs = db
    import_dataset(conn, dataset(tmp_path))
    conn.execute("INSERT INTO documents(id,source,chunker,tokenizer,max_tokens) VALUES ('private','s','c','t',512)")
    conn.execute('''INSERT INTO chunks SELECT 'private:0','private',0,headings,text,
                    contextualized_text,contextualized_tokens,embedding FROM chunks LIMIT 1''')
    original = httpx.AsyncClient
    monkeypatch.setattr('src.rag.httpx.AsyncClient', lambda **kw: original(**kw, transport=httpx.MockTransport(
        lambda _: httpx.Response(200, json={'documents': [], 'nextCursor': ''}))))

    async def run():
        service = RagService.__new__(RagService)
        service.settings = SimpleNamespace(note_url='http://note', indexer_service_token=SecretStr('token'), rag_top_k=6)
        service.pool = AsyncConnectionPool(conninfo=kwargs['conninfo'], kwargs={
            **{key: value for key, value in kwargs.items() if key != 'conninfo'},
            'autocommit': True, 'row_factory': dict_row},
                                          configure=register_vector_async, open=False)
        service.client = SimpleNamespace(generate=AsyncMock(return_value='Answer'))
        service.gate = asyncio.Semaphore(1)
        service.embed = lambda _: np.array([1.0] + [0.0] * 1023, dtype='float32')
        await service.pool.open()
        try:
            request = SimpleNamespace(history=[], message='Guide', userContext={})
            # Note owns all citation ACLs, including bundled company documents.
            # An empty/revoked allowlist must not send those documents to the LLM.
            denied = await service.answer(request, [])
            assert denied['sources'] == []
            assert '<context>\n\n</context>' in service.client.generate.call_args.args[-1]
            result = await service.answer(request, ['doc'])
            assert result['sources']
            assert {source['documentId'] for source in result['sources']} == {'doc'}
            await service.index_documents()
        finally:
            await service.pool.close()
    asyncio.run(run())
    assert counts(conn) == (1, 2)
    assert conn.execute('SELECT origin FROM documents').fetchone() == ('bundled_json',)
