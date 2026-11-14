import hashlib
import json
from pathlib import Path

import pytest

from src.knowledge import DEFAULT_DIRECTORY, load_dataset, write_manifest


def document(doc_id='doc', chunks=2):
    return {'document_id': doc_id, 'source': 'guide.md', 'chunker': 'test', 'tokenizer': 'test',
            'max_tokens': 512, 'embedding_model': 'Qwen/Qwen3-Embedding-0.6B', 'embedding_dimensions': 1024,
            'chunks': [{'chunk_id': f'{doc_id}:{i}', 'chunk_index': i, 'headings': ['Guide'],
                        'text': 'Guide text', 'contextualized_text': 'Guide text',
                        'contextualized_tokens': 2, 'embedding': [1.0] + [0.0] * 1023} for i in range(chunks)]}


def dataset(path, docs=None, version='v1'):
    path.mkdir(exist_ok=True)
    for old in path.glob('*.json'):
        old.unlink()
    for doc in docs if docs is not None else [document()]:
        (path / (doc['document_id'] + '.json')).write_text(json.dumps(doc))
    write_manifest(path, version)
    return path


def test_shipped_dataset_checksums_and_dimensions():
    manifest, checksum, docs = load_dataset(DEFAULT_DIRECTORY)
    assert manifest.visibility == 'company'
    assert len(docs) == 23
    assert sum(entry.chunks for entry, _ in docs) == 637
    assert len(checksum) == 64


def test_detects_tampered_missing_and_unlisted_files(tmp_path):
    path = dataset(tmp_path)
    file = path / 'doc.json'
    original = file.read_bytes()
    file.write_bytes(original + b' ')
    with pytest.raises(ValueError, match='Checksum'):
        load_dataset(path)
    file.unlink()
    with pytest.raises(ValueError, match='every dataset'):
        load_dataset(path)
    file.write_bytes(original)
    (path / 'extra.json').write_text('{}')
    with pytest.raises(ValueError, match='every dataset'):
        load_dataset(path)


@pytest.mark.parametrize('mutate', [
    lambda d: d['chunks'][0].update(embedding=[0.0] * 1024),
    lambda d: d['chunks'][0].update(embedding=[1.0] * 3),
    lambda d: d['chunks'][0]['embedding'].__setitem__(0, True),
    lambda d: d['chunks'][0]['embedding'].__setitem__(0, float('nan')),
    lambda d: d.update(embedding_model='another-model'),
    lambda d: d['chunks'][1].update(chunk_index=0),
    lambda d: d['chunks'][1].update(chunk_id=d['chunks'][0]['chunk_id']),
])
def test_rejects_invalid_data(tmp_path, mutate):
    path = dataset(tmp_path)
    data = json.loads((path / 'doc.json').read_text())
    mutate(data)
    content = json.dumps(data).encode()
    (path / 'doc.json').write_bytes(content)
    manifest = json.loads((path / 'manifest.json').read_text())
    manifest['files'][0]['sha256'] = hashlib.sha256(content).hexdigest()
    (path / 'manifest.json').write_text(json.dumps(manifest))
    with pytest.raises(ValueError):
        load_dataset(path)


def test_manifest_disallows_path_traversal_and_non_company_visibility(tmp_path):
    dataset(tmp_path)
    manifest = json.loads((tmp_path / 'manifest.json').read_text())
    manifest['files'][0]['file'] = '../secret.json'
    (tmp_path / 'manifest.json').write_text(json.dumps(manifest))
    with pytest.raises(ValueError):
        load_dataset(tmp_path)
    manifest['files'][0]['file'] = 'doc.json'
    manifest['visibility'] = 'private'
    (tmp_path / 'manifest.json').write_text(json.dumps(manifest))
    with pytest.raises(ValueError):
        load_dataset(tmp_path)
