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
        assert [m["role"] for m in payload["messages"]] == ["system", "user"]
        data = json.loads(payload["messages"][1]["content"].split("\n", 1)[1])
        assert data == {"question": "А что осталось?", "history": [
            {"role": "user", "content": "Посмотри мои заметки"},
            {"role": "assistant", "content": "Обычный текст вместо JSON"},
        ]}
        return httpx.Response(200, json={"choices": [{"message": {
            "content": '{"functions":["get_notes","get_checkpoints"]}'}}]})

    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")),
                            transport=httpx.MockTransport(handle))
        try:
            history = [SimpleNamespace(role="user", content="Посмотри мои заметки"),
                       SimpleNamespace(role="assistant", content="Обычный текст вместо JSON")]
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


def test_groq_retries_short_rate_limit_once():
    calls = []
    def handle(request):
        calls.append(request)
        if len(calls) == 1:
            return httpx.Response(429, headers={"retry-after": "0"}, json={"error": {"message": "quota"}})
        return httpx.Response(200, json={"choices": [{"message": {"content": "OK"}}]})
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")), transport=httpx.MockTransport(handle))
        try:
            assert await client.generate("Правила", [], "Вопрос") == "OK"
            assert len(calls) == 2
        finally:
            await client.close()
    asyncio.run(run())


def test_groq_does_not_wait_for_long_quota_limit():
    calls = []
    def handle(request):
        calls.append(request)
        return httpx.Response(429, headers={"retry-after": "3600"}, json={"error": {"message": "quota"}})
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")), transport=httpx.MockTransport(handle))
        try:
            with pytest.raises(httpx.HTTPStatusError):
                await client.plan([], "Вопрос")
            assert len(calls) == 1
        finally:
            await client.close()
    asyncio.run(run())


def test_groq_recovers_after_one_connect_timeout():
    calls = []
    def handle(request):
        calls.append(request)
        if len(calls) == 1:
            raise httpx.ConnectTimeout("TLS timeout", request=request)
        return httpx.Response(200, json={"choices": [{"message": {"content": "OK"}}]})
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")), transport=httpx.MockTransport(handle))
        try:
            assert await client.generate("Правила", [], "Вопрос") == "OK"
            assert len(calls) == 2
        finally:
            await client.close()
    asyncio.run(run())


def test_groq_does_not_retry_connect_timeout_forever():
    calls = []
    def handle(request):
        calls.append(request)
        raise httpx.ConnectTimeout("TLS timeout", request=request)
    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")), transport=httpx.MockTransport(handle))
        try:
            with pytest.raises(httpx.ConnectTimeout):
                await client.plan([], "Вопрос")
            assert len(calls) == 2
        finally:
            await client.close()
    asyncio.run(run())


@pytest.mark.parametrize("second_status", [200, 503])
def test_groq_retries_over_capacity_once(second_status, monkeypatch):
    calls = []
    waits = []

    async def sleep(delay):
        waits.append(delay)

    monkeypatch.setattr("src.llm.asyncio.sleep", sleep)

    def handle(request):
        calls.append(request)
        if len(calls) == 1 or second_status == 503:
            return httpx.Response(503, json={"error": {"message": "Model over capacity"}})
        return httpx.Response(200, json={"choices": [{"message": {"content": "OK"}}]})

    async def run():
        client = GroqClient(SimpleNamespace(groq_model="test-model", groq_api_key=SecretStr("key")), transport=httpx.MockTransport(handle))
        try:
            if second_status == 200:
                assert await client.generate("Правила", [], "Вопрос") == "OK"
            else:
                with pytest.raises(httpx.HTTPStatusError):
                    await client.generate("Правила", [], "Вопрос")
            assert len(calls) == 2
            assert waits == [1]
        finally:
            await client.close()

    asyncio.run(run())
