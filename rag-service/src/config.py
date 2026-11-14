from pathlib import Path

from pydantic import Field, SecretStr
from pydantic_settings import BaseSettings, SettingsConfigDict

ROOT = Path(__file__).resolve().parents[1]
EMBEDDING_MODEL = "Qwen/Qwen3-Embedding-0.6B"


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=ROOT / ".env", extra="ignore")

    postgres_host: str
    postgres_port: int = 5432
    postgres_user: str
    postgres_password: SecretStr
    postgres_db: str
    groq_api_key: SecretStr = Field(min_length=1)
    groq_model: str = Field(default="qwen/qwen3.8-27b", min_length=1)
    friend_token: SecretStr = Field(min_length=1)
    rag_top_k: int = Field(default=6, ge=1, le=20)
    embedding_max_tokens: int = Field(default=2048, ge=128, le=8192)
    note_url: str = "http://note_service:8079"
    indexer_service_token: SecretStr = Field(min_length=1)
    chat_service_token: SecretStr = Field(min_length=1)

    def db_kwargs(self):
        return dict(host=self.postgres_host, port=self.postgres_port,
                    user=self.postgres_user, password=self.postgres_password.get_secret_value(),
                    dbname=self.postgres_db, connect_timeout=10,
                    options="-c statement_timeout=15000")
