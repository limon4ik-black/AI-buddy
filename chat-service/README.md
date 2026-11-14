# Chat Service

Единственный сервис чата. Основан на новом REST/WebSocket-сервисе: UUID чатов, bigint ID сообщений, PostgreSQL. Совместимый REST API позволяет сохранить существующий интерфейс, источники, действия и повтор генерации.

Запуск — только через `make up` из корня репозитория. Настройки берутся из корневого `.env` и `deploy/compose.yaml`.

## API интерфейса (через Gateway)

| Метод | URL |
|---|---|
| GET, POST | /api/v1/chats |
| GET, PATCH, DELETE | /api/v1/chats/{id} |
| GET, POST | /api/v1/chats/{id}/messages |
| POST | /api/v1/chats/{id}/messages/{messageId}/answer |

POST messages принимает `{content, clientMessageId}`. Ответы используют camelCase; ID сообщений передаются строками, sequence — числом. Источники, suggestedAction, replyTo и mode сохраняются в БД. Повтор клиентского ключа не дублирует сообщение; повтор генерации возвращает сохранённый ответ.

## WebSocket и native REST

Gateway проксирует `/api/v1/chat-ws/*` в этот же процесс, без второго сервиса. Native API: POST/GET `/chats`, GET `/chats/{id}/messages`, DELETE `/chats/{id}`, GET `/ws`. Native JSON сохранён в snake_case. WS принимает `{chat_id, content}` и возвращает событие `{type, chat_id, role, content}` либо ошибку. REST интерфейса позволяет получить полные метаданные сохранённого ответа.

Нужен `Authorization: Bearer <JWT>`. Gateway передаёт сервисный токен, чат повторно проверяет JWT. Браузерный интерфейс использует REST; отдельного browser WS-клиента нет.

Оба транспорта получают контекст пользователя через Note Service и вызывают `/internal/v1/generate` RAG. Чужие документы и недоступные действия отбрасываются до сохранения ответа.

## Данные

Рабочие таблицы: `chat_db.chat_v2.chats` и `chat_db.chat_v2.messages`. Пользователи со строковыми JWT subject стабильно отображаются в UUID владельца. В API интерфейса userId остаётся исходным subject. Старые public-таблицы сохранены, но автоматически не импортируются.
