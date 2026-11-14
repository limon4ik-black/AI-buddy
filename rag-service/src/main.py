import logging
import secrets
from contextlib import asynccontextmanager
from typing import Annotated, Literal
from urllib.parse import quote

import httpx

from fastapi import Depends, FastAPI, HTTPException, Request, Header
from fastapi.security import HTTPAuthorizationCredentials, HTTPBearer
from pydantic import BaseModel, ConfigDict, Field, StringConstraints

from .config import Settings
from .runtime import single_instance

logger = logging.getLogger(__name__)
Text = Annotated[str, StringConstraints(strip_whitespace=True, min_length=1, max_length=8000)]
security = HTTPBearer(auto_error=False)


class Turn(BaseModel):
    model_config = ConfigDict(extra="forbid")
    role: Literal["user", "assistant"]
    content: Text


class ChatRequest(BaseModel):
    model_config = ConfigDict(extra="forbid")
    message: Text
    history: list[Turn] = Field(default_factory=list, max_length=20)


class Source(BaseModel):
    id: str
    source: str
    headings: list[str]
    score: float


class ChatResponse(BaseModel):
    answer: str
    sources: list[Source]


class GenerateRequest(ChatRequest):
    requestId: str
    userId: str = Field(min_length=1, max_length=255)
    userContext: dict


@asynccontextmanager
async def lifespan(app: FastAPI):
    with single_instance():
        from .rag import RagService

        app.state.settings = Settings()
        service = RagService(app.state.settings)
        try:
            await service.start()
            app.state.service = service
            yield
        finally:
            await service.close()


app = FastAPI(title="Avito AI Buddy", lifespan=lifespan)


@app.get("/healthz")
async def health():
    return {"status": "ok"}


async def fetch_user_context(user_id: str, authorization: str, request: Request) -> dict:
    settings = request.app.state.settings
    try:
        async with httpx.AsyncClient(
            timeout=httpx.Timeout(15, connect=5),
            transport=getattr(request.app.state, "info_transport", None),
        ) as client:
            response = await client.get(
                settings.note_url.rstrip("/") + "/internal/v1/users/" + quote(user_id, safe="") + "/context",
                headers={"Authorization": authorization,
                         "X-Service-Token": settings.chat_service_token.get_secret_value()},
            )
        if response.status_code in (401, 403):
            raise HTTPException(response.status_code, "Note Service запретил доступ к контексту пользователя")
        response.raise_for_status()
        context = response.json()
    except HTTPException:
        raise
    except (httpx.HTTPError, ValueError) as exc:
        logger.error("Info context failed: %s", type(exc).__name__)
        raise HTTPException(502, "Не удалось получить контекст из Note Service") from None
    profile = context.get("profile") if isinstance(context, dict) else None
    if not isinstance(profile, dict) or profile.get("userId") != user_id:
        raise HTTPException(502, "Note Service вернул контекст другого пользователя")
    allowed = context.get("allowedDocumentIds")
    tools = context.get("tools", [])
    if (not isinstance(allowed, list) or any(not isinstance(item, str) for item in allowed)
            or not isinstance(tools, list) or any(not isinstance(tool, dict) for tool in tools)):
        raise HTTPException(502, "Note Service вернул некорректный контекст")
    return context


@app.post("/internal/v1/generate")
async def generate(body: GenerateRequest, request: Request,
                   x_service_token: Annotated[str | None, Header()] = None,
                   authorization: Annotated[str | None, Header()] = None):
    token = request.app.state.settings.friend_token.get_secret_value()
    if not x_service_token or not secrets.compare_digest(x_service_token.encode(), token.encode()):
        raise HTTPException(401, "Неверный сервисный токен")
    profile = body.userContext.get("profile")
    if not isinstance(profile, dict) or not isinstance(profile.get("userId"), str) or not profile["userId"]:
        raise HTTPException(422, "Нужен авторизованный контекст")
    if body.userId != profile["userId"]:
        raise HTTPException(422, "userId должен совпадать с Auth ID в userContext.profile.userId")
    allowed = body.userContext.get("allowedDocumentIds", [])
    if not isinstance(allowed, list) or any(not isinstance(item, str) for item in allowed):
        raise HTTPException(422, "Некорректные разрешённые документы")
    scheme, _, bearer = (authorization or "").partition(" ")
    if scheme.lower() != "bearer" or not bearer or bearer.strip() != bearer:
        raise HTTPException(401, "Нужен JWT пользователя в Authorization: Bearer")
    # Info verifies JWT sub against the requested Auth ID. Never retrieve using
    # submitted context/ACL: resolve the authoritative context for this request.
    context = await fetch_user_context(body.userId, authorization, request)
    body = body.model_copy(update={"userContext": context})
    allowed = context["allowedDocumentIds"]
    try:
        result = await request.app.state.service.answer(body, allowed, authorization=authorization)
        action = None
        question = body.message.lower()
        if any(word in question for word in ("доступ", "подключ", "инструмент")):
            for tool in body.userContext.get("tools", []):
                if (tool.get("documentId") in allowed and
                    any(str(tool.get(key, "")).lower() in question
                        for key in ("id", "name") if tool.get(key))):
                    action = {"type": "REQUEST_ACCESS", "label": "Запросить доступ", "system": tool["id"]}
                    break
        return {"answer": result["answer"], "sources": [
            {"documentId": source["documentId"], "title": source["title"],
             "section": " / ".join(source["headings"])} for source in result["sources"]
        ], "suggestedAction": action, "mode": "rag"}
    except Exception as exc:
        logger.error("Generate failed: %s", type(exc).__name__)
        raise HTTPException(502, "Не удалось получить ответ") from None


def authenticate(request: Request, credentials: Annotated[
    HTTPAuthorizationCredentials | None, Depends(security)
]):
    token = request.app.state.settings.friend_token.get_secret_value()
    if credentials is None or not secrets.compare_digest(
        credentials.credentials.encode(), token.encode()
    ):
        raise HTTPException(401, "Неверный Bearer-токен", headers={"WWW-Authenticate": "Bearer"})


@app.post("/chat", response_model=ChatResponse, dependencies=[Depends(authenticate)])
async def chat(body: ChatRequest, request: Request):
    try:
        return await request.app.state.service.answer(body)
    except Exception as exc:
        # Avoid logging provider URLs / credentials / private conversation text.
        logger.error("Chat failed: %s", type(exc).__name__)
        raise HTTPException(502, "Не удалось получить ответ. Попробуйте позже.") from None
