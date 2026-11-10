from types import SimpleNamespace
from unittest.mock import AsyncMock

import pytest
from fastapi.testclient import TestClient
from pydantic import SecretStr

from src.main import app


@pytest.fixture
def client():
    app.state.settings = SimpleNamespace(friend_token=SecretStr("test-token"))
    app.state.service = SimpleNamespace(answer=AsyncMock(return_value={
        "answer": "Процесс описан в playbook [1]", "sources": []
    }))
    return TestClient(app)


def test_auth_required(client):
    for headers in ({}, {"Authorization": "Bearer wrong"}, {"Authorization": "Basic test-token"}):
        assert client.post("/chat", json={"message": "Вопрос"}, headers=headers).status_code == 401
    app.state.service.answer.assert_not_called()


def test_chat_passes_history(client):
    response = client.post("/chat", headers={"Authorization": "Bearer test-token"}, json={
        "message": "А разработка?", "history": [
            {"role": "user", "content": "Расскажи про Авито"},
            {"role": "assistant", "content": "Авито — это..."}
        ]
    })
    assert response.status_code == 200
    assert response.json()["answer"].endswith("[1]")
    body = app.state.service.answer.call_args.args[0]
    assert body.history[1].role == "assistant"
    assert body.message == "А разработка?"


@pytest.mark.parametrize("body", [
    {"message": " "}, {"message": "x", "history": [{"role": "system", "content": "x"}]},
    {"message": "x" * 8001}, {"message": "x", "history": [{"role": "user", "content": "x"}] * 21}
])
def test_validation(client, body):
    assert client.post("/chat", headers={"Authorization": "Bearer test-token"}, json=body).status_code == 422


def test_failure_does_not_expose_secrets(client):
    app.state.service.answer.side_effect = RuntimeError("secret-key")
    response = client.post("/chat", headers={"Authorization": "Bearer test-token"}, json={"message": "x"})
    assert response.status_code == 502
    assert "secret-key" not in response.text
