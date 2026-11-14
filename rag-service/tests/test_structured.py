import asyncio
from types import SimpleNamespace
from unittest.mock import AsyncMock

import httpx
import pytest

from src.structured import collect_data, get_notes


def request():
    return SimpleNamespace(message="Что осталось по моим записям?", history=[], userContext={
        "onboarding": {"tasks": [{"id": "task", "title": "Встреча", "status": "pending"}],
                       "total": 1, "completed": 0, "percent": 0}})


def test_only_selected_functions_and_forwarded_jwt():
    calls = []

    def handle(req):
        calls.append(req)
        assert req.method == "GET"
        assert req.url.path == "/api/v1/notes"
        assert "user_id" not in req.url.params
        assert req.headers["Authorization"] == "Bearer signed-user-jwt"
        return httpx.Response(200, json={"items": [{"id": "n", "content": "Запись", "userId": "private-id"}]})

    async def run():
        llm = SimpleNamespace(plan=AsyncMock(return_value=[]))
        args = (llm, request(), SimpleNamespace(note_url="http://note"), "Bearer signed-user-jwt")
        transport = httpx.MockTransport(handle)
        assert await collect_data(*args, transport=transport) == {}
        assert calls == []
        llm.plan.return_value = ["get_checkpoints"]
        result = await collect_data(*args, transport=transport)
        assert result["get_checkpoints"]["items"][0]["status"] == "pending"
        assert calls == []
        llm.plan.return_value = ["get_notes", "get_checkpoints", "get_notes"]
        result = await collect_data(*args, transport=transport)
        assert set(result) == {"get_notes", "get_checkpoints"}
        assert len(calls) == 1
        assert "userId" not in result["get_notes"]["items"][0]

    asyncio.run(run())


@pytest.mark.parametrize("failure", [401, 403, 500, "network", "invalid"])
def test_function_failure_is_not_empty_and_other_function_survives(failure):
    def handle(req):
        if failure == "network":
            raise httpx.ConnectError("private details", request=req)
        if failure == "invalid":
            return httpx.Response(200, json={"items": "bad"})
        return httpx.Response(failure, json={"error": "private details"})

    async def run():
        llm = SimpleNamespace(plan=AsyncMock(return_value=["get_notes", "get_checkpoints"]))
        result = await collect_data(llm, request(), SimpleNamespace(note_url="http://note"), "Bearer jwt",
                                    transport=httpx.MockTransport(handle))
        assert result["get_notes"] == {"status": "unavailable"}
        assert result["get_checkpoints"]["status"] == "ok"

    asyncio.run(run())


def test_notes_empty_and_partial_results_are_distinct():
    async def run(items):
        return await get_notes(SimpleNamespace(note_url="http://note"), "Bearer jwt",
            transport=httpx.MockTransport(lambda _: httpx.Response(200, json={"items": items})))

    assert asyncio.run(run([])) == {"status": "ok", "items": [], "truncated": False}
    result = asyncio.run(run([{"id": str(i), "content": "x" * 20000} for i in range(21)]))
    assert len(result["items"]) == 20
    assert result["truncated"] is True
    assert sum(len(item["content"]) for item in result["items"]) <= 16000
    assert result["items"][-1]["id"] == "19"
