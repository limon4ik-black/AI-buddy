import asyncio
import json
from types import SimpleNamespace

import httpx
import pytest
from pydantic import SecretStr

from src.llm import GroqClient


def test_planner_selects_functions_using_json_and_history():
    def handle(request):
        payload = json.loads(request.content)
        assert payload["response_format"] == {"type": "json_object"}
        assert payload["messages"][1]["content"] == "Посмотри мои заметки"
        return httpx.Response(200, json={"choices": [{"message": {
            "content": '{"functions":["get_notes","get_checkpoints"]}'}}]})

    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")),
                            transport=httpx.MockTransport(handle))
        try:
            history = [SimpleNamespace(role="user", content="Посмотри мои заметки")]
            assert await client.plan(history, "А что осталось?") == ["get_notes", "get_checkpoints"]
        finally:
            await client.close()

    asyncio.run(run())


@pytest.mark.parametrize("content", [
    '{"functions":["delete_notes"]}', '{"functions":[],"userId":"other"}',
    '{"functions":"get_notes"}', 'not json', '{}',
])
def test_planner_rejects_invalid_or_unknown_functions(content):
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")),
            transport=httpx.MockTransport(lambda _: httpx.Response(200, json={
                "choices": [{"message": {"content": content}}]})))
        try:
            with pytest.raises(ValueError):
                await client.plan([], "Вопрос")
        finally:
            await client.close()

    asyncio.run(run())


def test_groq_sends_history_context_and_auth():
    def handler(request):
        assert str(request.url) == "https://api.groq.com/openai/v1/chat/completions"
        assert request.headers["Authorization"] == "Bearer test-key"
        payload = json.loads(request.content)
        assert payload["model"] == "test-model"
        assert payload["messages"] == [
            {"role": "system", "content": "Правила"},
            {"role": "user", "content": "История"},
            {"role": "assistant", "content": "Ответ"},
            {"role": "user", "content": "Контекст [1] и вопрос"},
        ]
        return httpx.Response(200, json={"choices": [{"message": {"content": "Ответ [1]"}}]})

    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("test-key")),
                            transport=httpx.MockTransport(handler))
        try:
            history = [SimpleNamespace(role="user", content="История"),
                       SimpleNamespace(role="assistant", content="Ответ")]
            assert await client.generate("Правила", history, "Контекст [1] и вопрос") == "Ответ [1]"
        finally:
            await client.close()

    asyncio.run(run())


@pytest.mark.parametrize("status,content,error", [
    (429, {"error": {"message": "quota"}}, httpx.HTTPStatusError),
    (200, {"choices": [{"message": {"content": " "}}]}, ValueError),
])
def test_groq_rejects_error_and_empty_answer(status, content, error):
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("test-key")),
                            transport=httpx.MockTransport(lambda _: httpx.Response(status, json=content)))
        try:
            with pytest.raises(error):
                await client.generate("Правила", [], "Вопрос")
        finally:
            await client.close()

    asyncio.run(run())
