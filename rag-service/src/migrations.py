"""Versioned schema migrations, including databases created by the old init SQL."""
import hashlib
from pathlib import Path

MIGRATIONS = Path(__file__).resolve().parents[1] / "postgresql" / "migrations"
# Shared by schema migration, bundled import, and Note synchronization.
INDEX_LOCK = 624739182


def migrate(conn, directory=MIGRATIONS):
    files = sorted(directory.glob("*.sql"))
    if not files:
        raise RuntimeError("RAG schema migrations are missing")
    with conn.transaction():
        conn.execute("SELECT pg_advisory_xact_lock(%s)", (INDEX_LOCK,))
        conn.execute("""CREATE TABLE IF NOT EXISTS rag_schema_migrations (
            name TEXT PRIMARY KEY, sha256 TEXT NOT NULL,
            applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW())""")
        applied = dict(conn.execute("SELECT name, sha256 FROM rag_schema_migrations").fetchall())
        for path in files:
            content = path.read_bytes()
            checksum = hashlib.sha256(content).hexdigest()
            if path.name in applied:
                if applied[path.name] != checksum:
                    raise RuntimeError(f"Applied migration changed: {path.name}")
                continue
            conn.execute(content.decode("utf-8"))
            conn.execute("INSERT INTO rag_schema_migrations(name, sha256) VALUES (%s, %s)",
                         (path.name, checksum))
