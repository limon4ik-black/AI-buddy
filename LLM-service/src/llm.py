import httpx

from .config import Settings


class GroqClient:
    def __init__(self, settings: Settings, *, transport=None):
        self.model = settings.groq_model
        self.client = httpx.AsyncClient(
            base_url="https://api.groq.com/openai/v1/",
            headers={"Authorization": f"Bearer {settings.groq_api_key.get_secret_value()}"},
            timeout=httpx.Timeout(60, connect=10),
            transport=transport,
        )

    async def generate(self, system: str, history, prompt: str) -> str:
        messages = [{"role": "system", "content": system}]
        messages.extend({"role": turn.role, "content": turn.content} for turn in history)
        messages.append({"role": "user", "content": prompt})
        response = await self.client.post("chat/completions", json={
            "model": self.model, "messages": messages,
            "temperature": 0.2, "max_completion_tokens": 4096,
        })
        response.raise_for_status()
        answer = response.json()["choices"][0]["message"]["content"]
        if not isinstance(answer, str) or not answer.strip():
            raise ValueError("Groq вернул пустой ответ")
        return answer

    async def close(self):
        await self.client.aclose()
