"""Read-only personal data selected by the planner; identity always comes from JWT."""
import asyncio
import logging

import httpx

logger = logging.getLogger(__name__)
MAX_ITEMS = 20
MAX_TEXT = 4000
MAX_TOTAL_TEXT = 16000


def bounded_items(items, fields):
    if not isinstance(items, list) or any(not isinstance(item, dict) for item in items):
        raise ValueError("Invalid personal data")
    result = []
    truncated = len(items) > MAX_ITEMS
    budget = MAX_TOTAL_TEXT
    for item in items[:MAX_ITEMS]:
        entry = {}
        for field in fields:
            value = item.get(field)
            if isinstance(value, str):
                is_text = field in ("content", "title", "description")
                limit = min(MAX_TEXT, budget) if is_text else 256
                truncated |= len(value) > limit
                value = value[:limit]
                if is_text:
                    budget -= len(value)
            elif value is not None and not isinstance(value, (int, float, bool)):
                raise ValueError("Invalid personal data field")
            entry[field] = value
        result.append(entry)
    return {"status": "ok", "items": result, "truncated": truncated}


async def get_notes(settings, authorization, *, transport=None):
    async with httpx.AsyncClient(timeout=httpx.Timeout(15, connect=5), transport=transport) as client:
        response = await client.get(
            settings.note_url.rstrip("/") + "/api/v1/notes",
            headers={"Authorization": authorization},
            params={"limit": MAX_ITEMS + 1, "offset": 0},
        )
        response.raise_for_status()
        return bounded_items(response.json()["items"], ("id", "content", "createdAt", "updatedAt"))


async def get_checkpoints(context):
    # Snapshot fetched from Note using the verified user's JWT for this request.
    checklist = context["onboarding"]
    result = bounded_items(checklist["tasks"],
                           ("id", "title", "description", "period", "status", "completedAt"))
    result["summary"] = {key: checklist.get(key) for key in ("total", "completed", "percent")}
    return result


async def collect_data(llm, request, settings, authorization, *, transport=None):
    functions = await llm.plan(request.history, request.message)

    async def run(name):
        try:
            if name == "get_notes":
                return await get_notes(settings, authorization, transport=transport)
            if name == "get_checkpoints":
                return await get_checkpoints(request.userContext)
            raise ValueError("Unknown function")
        except (httpx.HTTPError, ValueError, KeyError, TypeError) as exc:
            logger.warning("Personal data function %s failed: %s", name, type(exc).__name__)
            return {"status": "unavailable"}

    names = list(dict.fromkeys(functions))
    return dict(zip(names, await asyncio.gather(*(run(name) for name in names))))
