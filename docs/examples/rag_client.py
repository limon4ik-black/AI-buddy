"""Read-only client for the ML engineer. No third-party dependencies.
INDEXER_SERVICE_TOKEN=... NOTE_URL=http://localhost:8079 python3 docs/examples/rag_client.py
The output includes ACL metadata. Filter before retrieval, not after generation.
"""
import json
import os
import urllib.parse
import urllib.request


def snapshot(base_url, service_token):
    def get(path):
        request = urllib.request.Request(base_url.rstrip('/') + path, headers={'X-Service-Token': service_token})
        with urllib.request.urlopen(request, timeout=10) as response:
            return json.load(response)

    teams = get('/internal/v1/teams')['items']
    details = [get('/internal/v1/teams/' + urllib.parse.quote(t['id'], safe='')) for t in teams]
    services = get('/internal/v1/services')['items']
    templates = [get('/internal/v1/checklist/templates?teamId=' + urllib.parse.quote(t['id'], safe='')) for t in teams]
    documents, cursor = [], ''
    while True:
        page = get('/internal/v1/knowledge/documents?' + urllib.parse.urlencode({'after': cursor, 'limit': 100}))
        documents.extend(page['documents'])
        cursor = page['nextCursor']
        if not cursor:
            break
    return {'schemaVersion': '1.1.0', 'teams': details, 'services': services, 'checklistTemplates': templates, 'documents': documents}


if __name__ == '__main__':
    token = os.environ.get('INDEXER_SERVICE_TOKEN')
    if not token:
        raise SystemExit('Set INDEXER_SERVICE_TOKEN (server-side only)')
    print(json.dumps(snapshot(os.environ.get('NOTE_URL', 'http://localhost:8079'), token), ensure_ascii=False, indent=2))
