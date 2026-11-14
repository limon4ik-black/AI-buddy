# RAG Service

Запуск из корня: `make rag`. Нужен `GROQ_API_KEY` в корневом `.env`; модель генерации задаётся `GROQ_MODEL`.

Note Service → Qwen3-Embedding-0.6B → PostgreSQL/pgvector → Groq. Embeddings имеют размерность1024. При наличии CUDA используется GPU/FP16, иначе CPU/FP32. Модель скачивается при первом запуске и сохраняется в volume.

При каждом старте сервис загружает полный снимок документов Note Service, синхронизирует изменения и удаляет отсутствующие документы одной транзакцией. Неизменённые документы повторно не кодируются. Пока процесс работает, автоматического обновления индекса нет: после изменения документов перезапустите rag.

Основной контракт: POST `/internal/v1/generate`, заголовок `X-Service-Token` со значением RAG_SERVICE_TOKEN. Тело: `{requestId, userId, message, history, userContext}`. userId — Auth users.id / JWT sub, обязан совпадать с userContext.profile.userId. UUID владельца из БД чата сюда не передаётся. Несовпадение отклоняется с422 до retrieval. В userContext обязательны profile.userId и allowedDocumentIds. Поиск ограничивается разрешёнными ID до генерации. Ответ: `{answer, sources: [{documentId,title,section}], suggestedAction, mode: "rag"}`.

Старый POST `/chat` с Bearer FRIEND_TOKEN сохранён для совместимости тестов и внутренних клиентов; в общем приложении пользовательские запросы идут только через `/internal/v1/generate`. Не публикуйте сервис напрямую наружу.

GET `/healthz` доступен после загрузки модели и синхронизации индекса. На CPU первый старт может занимать несколько минут.

Для `/internal/v1/generate` также нужен исходный пользовательский JWT в `Authorization: Bearer ...`. До поиска RAG вызывает Info `GET /internal/v1/users/{userId}/context` с тем же JWT и CHAT_SERVICE_TOKEN. Контекст и ACL из Info авторитетны; переданный клиентом контекст не расширяет доступ. Публичный userId всегда равен Auth ID, а не UUID владельца из БД чата.
