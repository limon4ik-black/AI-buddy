import logging
import secrets
from contextlib import asynccontextmanager
from typing import Annotated, Literal

from fastapi import Depends, FastAPI, HTTPException, Request
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
