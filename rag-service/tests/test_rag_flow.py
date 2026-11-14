import asyncio
import json
from contextlib import asynccontextmanager
from types import SimpleNamespace
from unittest.mock import AsyncMock

import httpx
import pytest
from pydantic import SecretStr

from src.llm import GroqClient
from src.main import GenerateRequest
from src.rag import RagService


@pytest.mark.parametrize('functions', [[], ['get_notes'], ['get_checkpoints'], ['get_notes', 'get_checkpoints']])
def test_two_stage_answer_passes_only_selected_data_and_preserves_acl(monkeypatch, functions):
    calls = []
    notes = AsyncMock(return_value={'status': 'ok', 'items': [{'content': 'NOTE_MARKER'}], 'truncated': False})
    monkeypatch.setattr('src.structured.get_notes', notes)

    def handle(req):
        payload = json.loads(req.content)
        calls.append(payload)
        if len(calls) == 1:
            return httpx.Response(200, json={'choices': [{'message': {
                'content': json.dumps({'functions': functions})}}]})
        prompt = payload['messages'][-1]['content']
        assert ('NOTE_MARKER' in prompt) == ('get_notes' in functions)
        assert ('CHECKPOINT_MARKER' in prompt) == ('get_checkpoints' in functions)
        assert 'KNOWLEDGE_MARKER' in prompt
        return httpx.Response(200, json={'choices': [{'message': {'content': 'Ответ [1]'}}]})

    cursor = SimpleNamespace(fetchall=AsyncMock(return_value=[{
        'id': 'chunk', 'document_id': 'guide', 'source': 'guide-url', 'headings': ['Guide'],
        'contextualized_text': 'KNOWLEDGE_MARKER', 'score': 0.9}]))
    connection = SimpleNamespace(execute=AsyncMock(return_value=cursor))

    @asynccontextmanager
    async def connect():
        yield connection

    async def run():
        service = RagService.__new__(RagService)
        service.settings = SimpleNamespace(groq_model='model', groq_api_key=SecretStr('key'), rag_top_k=6)
        service.client = GroqClient(service.settings, transport=httpx.MockTransport(handle))
        service.gate = asyncio.Semaphore(1)
        service.embed = lambda _: [1.0]
        service.pool = SimpleNamespace(connection=connect)
        body = GenerateRequest(requestId='r', userId='user', message='Что дальше?', userContext={
            'profile': {'userId': 'user'}, 'allowedDocumentIds': ['guide'],
            'onboarding': {'tasks': [{'id': 't', 'title': 'CHECKPOINT_MARKER', 'status': 'pending'}],
                           'total': 1, 'completed': 0, 'percent': 0}})
        try:
            result = await service.answer(body, ['guide'], authorization='Bearer jwt')
            assert result['answer'] == 'Ответ [1]'
            assert result['sources'][0]['documentId'] == 'guide'
            assert len(calls) == 2
            assert connection.execute.call_args.args[1][2:4] == (['guide'], ['guide'])
            if 'get_notes' in functions:
                assert notes.await_args.args[1] == 'Bearer jwt'
            else:
                notes.assert_not_awaited()
        finally:
            await service.client.close()

    asyncio.run(run())
