"""Validate OpenAPI documents and actual HTTP responses captured by integration tests."""
import json
import re
from pathlib import Path
import yaml
from openapi_spec_validator import validate_spec
from openapi_schema_validator import OAS30Validator
root = Path(__file__).resolve().parents[1]
specs = {}
for file in sorted((root / 'api').glob('openapi-*.yaml')):
    spec = yaml.safe_load(file.read_text())
    validate_spec(spec)
    specs[file.stem.removeprefix('openapi-')] = spec
    print(f'Valid OpenAPI 3.0: {file.name}')
fixtures = json.loads((root / 'reports/contract-fixtures.json').read_text())
validated = 0
for f in fixtures:
    spec = specs[f['spec']]
    found = None
    for route, methods in spec['paths'].items():
        pattern = re.sub(r'\{[^}]+\}', '[^/]+', route)
        if re.fullmatch(pattern, f['path']) and f['method'] in methods:
            found = methods[f['method']]
            break
    if found is None:
        continue  # Deliberate requests to nonexistent routes.
    response = found['responses'].get(str(f['status']))
    assert response, f"Undocumented status: {f['method']} {f['path']} {f['status']}"
    if 'content' not in response:
        continue
    schema = {**response['content']['application/json']['schema'], 'components': spec['components']}
    OAS30Validator(schema).validate(f['body'])
    validated += 1
print(f'Validated {validated} real HTTP responses against OpenAPI schemas')
assert validated >= 100, 'Run the integration tests to capture real responses first'
