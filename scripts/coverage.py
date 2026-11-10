"""Count source lines mapped to Go cover blocks and combine with V8 line coverage.
Go's official statement report is kept separately. No application file is excluded.
Blank lines, comments, and brace-only lines are not executable source lines.
"""
import json
import re
from pathlib import Path
root = Path(__file__).resolve().parents[1]
line_hits = {}
for record in (root / 'reports/backend.cover').read_text().splitlines()[1:]:
    location, statements, hits = record.split()
    filename, span = location.split(':')
    filename = filename.split('/backend/', 1)[1]
    start, end = [int(part.split('.')[0]) for part in span.split(',')]
    lines = (root / 'backend' / filename).read_text().splitlines()
    for number in range(start, end + 1):
        text = lines[number - 1].strip()
        if not text or text.startswith('//') or not re.search(r'[A-Za-zА-Яа-я0-9_]', text):
            continue
        key = (filename, number)
        line_hits[key] = max(line_hits.get(key, 0), int(hits) > 0)
backend = {'total': len(line_hits), 'covered': sum(line_hits.values())}
frontend = json.loads((root / 'reports/frontend/coverage-summary.json').read_text())['total']['lines']
combined = {k: backend[k] + frontend[k] for k in ('total', 'covered')}
for entry in (backend, combined):
    entry['pct'] = round(entry['covered'] * 100 / entry['total'], 2)
report = {'method': 'Go cover mapped source lines + Vitest V8 line coverage; all Go cmd/internal files and production frontend TS/TSX files are included; test helpers are excluded.', 'backendLines': backend, 'frontendLines': frontend, 'combinedLines': combined}
(root / 'reports/coverage-summary.json').write_text(json.dumps(report, ensure_ascii=False, indent=2) + '\n')
print(json.dumps(report, ensure_ascii=False, indent=2))
assert backend['pct'] > 80 and frontend['pct'] > 80 and combined['pct'] > 80, 'Line coverage must exceed 80%'
