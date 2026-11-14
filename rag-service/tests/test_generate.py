from copy import deepcopy
from types import SimpleNamespace
from unittest.mock import AsyncMock

import httpx
import pytest
from fastapi.testclient import TestClient
from pydantic import SecretStr

from src.main import app

HEADERS = {"X-Service-Token": "service-token", "Authorization": "Bearer user-jwt"}


def context():
    return {"profile": {"userId": "intern-egor", "fullName": "Егор из Info"},
            "allowedDocumentIds": ["guide"],
            "tools": [{"id": "stash", "name": "Stash", "documentId": "guide"}]}


def prepare(info_context=None, status=200, failure=False):
    app.state.settings = SimpleNamespace(
        friend_token=SecretStr("service-token"), note_url="http://note_service:8079",
        chat_service_token=SecretStr("chat-service-token"),
    )
    app.state.info_requests = []

    def info(request):
        app.state.info_requests.append(request)
        assert request.method == "GET"
        assert str(request.url) == "http://note_service:8079/internal/v1/users/intern-egor/context"
        assert request.headers["Authorization"] == "Bearer user-jwt"
        assert request.headers["X-Service-Token"] == "chat-service-token"
        if failure:
            raise httpx.ConnectError("private-upstream-details", request=request)
        return httpx.Response(status, json=context() if info_context is None else info_context)

    app.state.info_transport = httpx.MockTransport(info)
    app.state.service = SimpleNamespace(answer=AsyncMock(return_value={
        "answer": "Read the guide", "sources": [{
            "documentId": "guide", "title": "Guide", "headings": ["Setup"],
        }],
    }))
    return TestClient(app)


def payload():
    return {"requestId": "request", "userId": "intern-egor", "message": "Дай доступ в Stash", "history": [],
            "userContext": context()}


def test_generate_requires_service_token():
    client = prepare()
    assert client.post("/internal/v1/generate", json=payload()).status_code == 401
    app.state.service.answer.assert_not_called()
    assert app.state.info_requests == []


def test_generate_fetches_info_with_jwt_and_returns_chat_contract():
    client = prepare()
    response = client.post("/internal/v1/generate", json=payload(), headers=HEADERS)
    assert response.status_code == 200
    assert len(app.state.info_requests) == 1
    assert app.state.service.answer.call_args.args[1] == ["guide"]
    assert response.json() == {
        "answer": "Read the guide", "mode": "rag",
        "sources": [{"documentId": "guide", "title": "Guide", "section": "Setup"}],
        "suggestedAction": {"type": "REQUEST_ACCESS", "label": "Запросить доступ", "system": "stash"},
    }


def test_generate_uses_authoritative_info_acl_and_profile():
    trusted = context()
    trusted["tools"] = []
    client = prepare(trusted)
    body = payload()
    body["userContext"]["allowedDocumentIds"].append("another-teams-secret")
    body["userContext"]["profile"]["fullName"] = "Untrusted name"
    response = client.post("/internal/v1/generate", json=body, headers=HEADERS)
    assert response.status_code == 200
    passed_body, allowed = app.state.service.answer.call_args.args
    assert allowed == ["guide"]
    assert passed_body.userContext == trusted
    assert response.json()["suggestedAction"] is None


@pytest.mark.parametrize("authorization", [None, "Basic user-jwt", "Bearer "])
def test_generate_requires_forwarded_user_jwt(authorization):
    client = prepare()
    headers = {"X-Service-Token": "service-token"}
    if authorization is not None:
        headers["Authorization"] = authorization
    assert client.post("/internal/v1/generate", json=payload(), headers=headers).status_code == 401
    app.state.service.answer.assert_not_called()
    assert app.state.info_requests == []


@pytest.mark.parametrize("upstream_status,expected", [(401, 401), (403, 403), (500, 502), (404, 502)])
def test_info_failure_blocks_retrieval(upstream_status, expected):
    client = prepare(status=upstream_status)
    response = client.post("/internal/v1/generate", json=payload(), headers=HEADERS)
    assert response.status_code == expected
    app.state.service.answer.assert_not_called()


def test_info_network_failure_does_not_expose_details():
    client = prepare(failure=True)
    response = client.post("/internal/v1/generate", json=payload(), headers=HEADERS)
    assert response.status_code == 502
    assert "private-upstream-details" not in response.text
    app.state.service.answer.assert_not_called()


def test_info_mismatched_identity_blocks_retrieval():
    trusted = context()
    trusted["profile"]["userId"] = "another-user"
    client = prepare(trusted)
    assert client.post("/internal/v1/generate", json=payload(), headers=HEADERS).status_code == 502
    app.state.service.answer.assert_not_called()


@pytest.mark.parametrize("trusted", [{}, {"profile": None}, {"profile": {"userId": "intern-egor"}, "allowedDocumentIds": "all"}])
def test_info_invalid_context_blocks_retrieval(trusted):
    client = prepare(deepcopy(trusted))
    assert client.post("/internal/v1/generate", json=payload(), headers=HEADERS).status_code == 502
    app.state.service.answer.assert_not_called()


def test_generate_rejects_invalid_acl_before_search():
    client = prepare()
    body = payload()
    body["userContext"]["allowedDocumentIds"] = "guide"
    assert client.post("/internal/v1/generate", json=body, headers=HEADERS).status_code == 422
    app.state.service.answer.assert_not_called()


def test_generate_rejects_different_profile_identity_before_search():
    client = prepare()
    body = payload()
    body["userId"] = "chat-storage-uuid"
    assert client.post("/internal/v1/generate", json=body, headers=HEADERS).status_code == 422
    app.state.service.answer.assert_not_called()
    assert app.state.info_requests == []


def test_generate_requires_canonical_user_id():
    client = prepare()
    body = payload()
    del body["userId"]
    assert client.post("/internal/v1/generate", json=body, headers=HEADERS).status_code == 422
    app.state.service.answer.assert_not_called()


@pytest.mark.parametrize("profile", [None, "wrong", [], {"userId": 42}])
def test_generate_rejects_malformed_profile_before_search(profile):
    client = prepare()
    body = payload()
    body["userContext"]["profile"] = profile
    assert client.post("/internal/v1/generate", json=body, headers=HEADERS).status_code == 422
    app.state.service.answer.assert_not_called()
