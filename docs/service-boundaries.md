# Employee и Note

Employee владеет профилями, людьми, командами, техническими сервисами, инструментами и заявками. Его существующая БД `info_db` сохранена с прежними именем и томом, чтобы обновление не создало пустую базу.

Note владеет `note.notes`, `board.boards`, `onboarding.knowledge_documents`, `onboarding.checklist_tasks` и `onboarding.user_task_progress` в отдельной БД `notes`. Межсервисных внешних ключей нет. В прогрессе user_id — текстовый Auth ID, команда берётся из Employee по JWT.

```text
RAG → Note → Employee (профиль, команда, инструменты)
       └──→ собственная БД (документы, чек-лист, прогресс)
```

RAG использует только NOTE_URL. Персональный контекст: `GET /internal/v1/users/{userId}/context`, CHAT_SERVICE_TOKEN + исходный JWT. Индексация: `GET /internal/v1/knowledge/documents`, INDEXER_SERVICE_TOKEN. Остальные справочные GET-ручки для индексатора доступны через Note, который проксирует каталоги Employee.

Gateway направляет `/api/v1/profile`, `/teams`, `/services`, `/tools`, `/access-requests` в Employee; `/checklist` и `/information` — в Note. `/notes-api` и `/board-api` продолжают работать как прежде.

## Перенос существующих данных

1. Employee применяет прежние миграции с прежним ключом `info` и подготавливает демо-данные.
2. note_migrate применяет SQL003 в БД notes.
3. note_import переносит документы, задачи и прогресс одной транзакцией из прежней БД. Согласованный снимок источника и текстовые user_id сохраняются.
4. Маркер `onboarding.imports/legacy-info-v1` запрещает повторное заполнение: новые изменения Note не перезаписываются при рестарте.
5. Только после успешного импорта запускается Note, затем RAG.

Исходные таблицы чек-листов и документов в info_db оставлены как резерв, но Employee больше не обслуживает их runtime-API. Источник доступен только отдельному процессу миграции; Note и RAG не получают DSN этой базы.

Привязка старого демо-прогресса к фактическому Auth ID выполняется Note один раз по подписанному email. Маркер identity_links и ON CONFLICT защищают последующие пользовательские изменения. Нельзя использовать people.id или внутренний UUID владельца чата вместо Auth ID.
