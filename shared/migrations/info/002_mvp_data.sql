-- Synthetic MVP data. These teams and procedures are not Avito regulations.

ALTER TABLE team_members DROP CONSTRAINT IF EXISTS team_members_role_check;

ALTER TABLE team_members ADD CONSTRAINT team_members_role_check CHECK(role IN ('lead','buddy','engineer','intern'));

UPDATE teams SET name='Backend',description='Развиваем серверные API, поиск и профиль пользователя. Работаем с Go, PostgreSQL и наблюдаемостью.' WHERE id='backend';

UPDATE teams SET name='Analytics',description='Помогаем командам принимать решения: метрики продукта, датасеты и аналитические исследования.' WHERE id='analytics';

INSERT INTO teams(id,name,description,buddy_id) VALUES
 ('frontend','Frontend','Создаём веб-интерфейсы, компоненты и доступный пользовательский опыт. React, TypeScript и дизайн-система.','polina'),
 ('qa','QA','Помогаем выпускать надёжные изменения: тест-дизайн, автоматизация и проверка релизов.','elena') ON CONFLICT DO NOTHING;

INSERT INTO people(id,full_name,position) VALUES
 ('polina','Полина Миронова','Frontend-разработчик'),
 ('nikita','Никита Белов','Тимлид Frontend'),
 ('danil','Данил Крылов','UI-разработчик'),
 ('elena','Елена Васильева','QA Automation Engineer'),
 ('andrey','Андрей Фомин','Тимлид QA'),
 ('roman','Роман Котов','Инженер по тестированию') ON CONFLICT DO NOTHING;

INSERT INTO team_members(team_id,person_id,role) VALUES
 ('frontend','polina','buddy'),
 ('frontend','nikita','lead'),
 ('frontend','danil','engineer'),
 ('qa','elena','buddy'),
 ('qa','andrey','lead'),
 ('qa','roman','engineer') ON CONFLICT DO NOTHING;

INSERT INTO technical_services(id,name,description,owner_team_id,technical_owner_id,lifecycle,repository_url,documentation_url) VALUES
 ('search-api','Search API','Поиск по учебному каталогу и фильтрация результатов.','backend','ivan','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('notification-api','Notification API','Демо-уведомления о событиях адаптации.','backend','alex','development','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('web-platform','Web Platform','Общий каркас веб-приложения и клиентская навигация.','frontend','nikita','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('design-system','Design System','Библиотека UI-компонентов и правила доступности.','frontend','polina','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('intern-portal','Intern Portal','Веб-интерфейс личного кабинета стажёра.','frontend','danil','development','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('metrics-catalog','Metrics Catalog','Определения продуктовых метрик и их владельцы.','analytics','daria','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('data-sandbox','Data Sandbox','Синтетические наборы данных для учебных исследований.','analytics','sasha','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('test-runner','Test Runner','Запуск автоматических проверок учебных проектов.','qa','elena','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('release-checks','Release Checks','Чек-листы smoke-проверок и фиксация результатов.','qa','roman','development','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy') ON CONFLICT DO NOTHING;

INSERT INTO knowledge_documents(id,team_id,title,content,category,visibility,source_url,synthetic) VALUES
 ('company-map',NULL,'Как устроена демо-компания','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Backend отвечает за серверные API. Frontend создаёт веб-интерфейсы. Analytics работает с метриками и учебными датасетами. QA помогает проверять изменения. В справочнике «Сервисы» указаны команда-владелец и конкретный ответственный. Если вопрос на стыке команд, начни со своего бадди.','company','company','https://github.com/avito-tech/playbook',true),
 ('access-company',NULL,'Общие доступы: Confluence и VPN','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Открой чат и попроси доступ к Confluence или VPN. Помощник предложит заявку. После твоего подтверждения тестовый сервис вернёт её номер и статус submitted. Он означает, что обращение зарегистрировано; доступ ещё не выдан. Согласование обсуди со своим бадди, после фактической проверки обнови чек-лист.','access','company','https://github.com/avito-tech/playbook',true),
 ('communications',NULL,'Чаты, встречи и рабочие договорённости','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Попроси бадди добавить тебя в основной канал команды и чат стажёров. Перед встречей прочитай повестку, запиши вопросы и уточни, где фиксируются решения. Если застрял на задаче, расскажи, что проверил, приложи безопасный пример и сформулируй, какая помощь нужна.','process','company','https://github.com/avito-tech/playbook',true),
 ('buddy-meetings',NULL,'Как подготовиться к 1:1 с бадди','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Договорись о регулярной встрече на 20–30 минут. Подготовь три пункта: что получилось, что мешает и что попробовать дальше. На первой встрече обсуди ожидания на месяц, критерии готовности первой задачи и удобный канал для коротких вопросов.','onboarding','company','https://github.com/avito-tech/playbook',true),
 ('security-basics',NULL,'Безопасная работа с учебными данными','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

В демо используются только вымышленные люди и тестовые инструкции. Не присылай пароли, токены, SSH-ключи и реальные данные сотрудников в чат. Секреты храни в локальной конфигурации, исключённой из Git. Доступ к инструкции своей команды не означает доступ к закрытым материалам других команд.','security','company','https://github.com/avito-tech/playbook',true),
 ('learning-plan',NULL,'Личный план обучения','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Выбери с бадди две технические темы и один навык взаимодействия. Для каждой темы запиши небольшой результат: прочитанный материал, учебный пример или короткое обсуждение. На второй неделе сверь план с первой задачей. Приоритеты можно менять после обратной связи.','learning','company','https://github.com/avito-tech/playbook',true),
 ('feedback',NULL,'Первая обратная связь','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Перед встречей собери примеры выполненных шагов и вопросы. Попроси бадди назвать одну сильную сторону и один следующий навык. Зафиксируй договорённости в своём плане и уточни, по какому результату будет видно продвижение.','onboarding','company','https://github.com/avito-tech/playbook',true),
 ('glossary',NULL,'Словарь новичка','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Бадди — коллега, который помогает адаптироваться. Тимлид — руководитель технической команды. Code review — обсуждение изменения кода до объединения. Smoke-проверка — короткая проверка основного пути. Submitted — заявка зарегистрирована, а не согласована.','company','company','https://github.com/avito-tech/playbook',true),
 ('backend-local','backend','Локальное окружение Go','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Проверь версию Go, загрузи зависимости из go.mod и запусти тесты учебного сервиса. Для общей демо-системы используй инструкцию проекта. После изменения контракта сверь OpenAPI. Если локальный порт занят, сначала найди процесс, а не меняй случайным образом настройки всей команды.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('backend-first-task','backend','Первая задача Backend','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Начни с небольшого изменения: валидация входа, исправление ошибки или дополнительный фильтр. Согласуй с бадди ожидаемое поведение и отрицательный сценарий. Покажи воспроизводимый пример, прогони тесты и попроси review у владельца сервиса.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('backend-review','backend','Как подготовить изменение к review','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

В описании изменения укажи проблему, новое поведение и способ проверки. Не объединяй несвязанные изменения. В ответ на замечание сначала уточни причину, затем предложи решение. Автор отвечает за понятность изменения и воспроизводимость тестов.','process','team','https://github.com/avito-tech/playbook',true),
 ('backend-observability','backend','Логи, метрики и обработка сбоев','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Найди healthcheck своего сервиса и проследи один запрос по requestId. Проверь, как возвращается ошибка зависимости. В логах не должно быть паролей и персональных сведений. Обсуди с бадди, какие метрики покажут отказ или замедление основного сценария.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('frontend-access','frontend','Доступы Frontend: Figma и Storybook','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

В Figma находятся учебные макеты, а в Storybook — состояния компонентов. Попроси доступ к нужному инструменту в чате и подтверди заявку. Согласование поможет проверить Полина. Номер заявки появится в профиле; он не означает автоматическую выдачу доступа.','access','team','https://github.com/avito-tech/playbook',true),
 ('frontend-local','frontend','Локальное окружение Frontend','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Используй версию Node.js, указанную в проекте, и установку по lock-файлу. Запусти dev-сервер и production-сборку. Проверь локальный API-клиент и обработку ошибок сети. Изучи структуру компонентов, прежде чем добавлять новую библиотеку.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('frontend-components','frontend','Компоненты и доступность','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Проверь состояния loading, empty, error и success. Интерактивные элементы должны быть доступны с клавиатуры и иметь понятные названия. Используй общие компоненты и дизайн-токены. Не передавай смысл только цветом: добавь текст или значок.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('frontend-first-task','frontend','Первая задача Frontend','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Возьми небольшую доработку экрана или компонента. Согласуй состояния и мобильное поведение с бадди. Проверь основной пользовательский путь, сохранение введённых данных после ошибки и отсутствие горизонтального скролла. Приложи результат проверки к review.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('frontend-review','frontend','Review веб-интерфейса','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Сверь результат с макетом, проверь длинный текст, пустые списки и узкий экран. Убедись, что состояние не смешивается между пользователями. Открой экран с клавиатуры и проверь названия кнопок. Для сетевых действий показывай ожидание и возможность повторить попытку.','process','team','https://github.com/avito-tech/playbook',true),
 ('analytics-access','analytics','Доступы Analytics: Metabase и Jupyter','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Metabase используется для учебных отчётов, Jupyter — для исследований на синтетических данных. Запроси инструмент через помощника и подтверди отправку. Номер заявки сохранится в профиле. Саша поможет уточнить согласование; реальные данные в демо не используются.','access','team','https://github.com/avito-tech/playbook',true),
 ('analytics-metrics','analytics','Словарь метрик','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Для каждой метрики уточни числитель, знаменатель, период, фильтры и владельца. Отличай отсутствие данных от нулевого результата. Перед сравнением проверь, что группы и определения совпадают. Не объявляй рост причинным эффектом без подходящей проверки.','analytics','team','https://github.com/avito-tech/playbook',true),
 ('analytics-notebook','analytics','Первый учебный notebook','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Открой синтетический датасет, опиши его строки и поля. Проверь пропуски и дубликаты. Сформулируй один вопрос и подготовь простой воспроизводимый расчёт. Подпиши график и вывод так, чтобы коллега мог понять их без устного пояснения.','analytics','team','https://github.com/avito-tech/playbook',true),
 ('analytics-first-task','analytics','Первая задача аналитика','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Начни с проверки определения метрики или небольшого учебного отчёта. Согласуй ожидаемый результат с заказчиком и бадди. Покажи ограничения данных, воспроизводимый запрос и следующий вопрос. Не используй настоящие персональные или финансовые сведения.','analytics','team','https://github.com/avito-tech/playbook',true),
 ('qa-access','qa','Доступы QA: TestRail и тестовый стенд','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

TestRail хранит учебные проверки, а тестовый стенд позволяет пройти основной сценарий. Попроси нужный доступ в чате и подтверди отправку. Елена поможет с согласованием. Все аккаунты и данные стенда демонстрационные, настоящие секреты в чат не передаются.','access','team','https://github.com/avito-tech/playbook',true),
 ('qa-local','qa','Окружение QA и тестовые данные','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Получи адрес учебного стенда и список тестовых аккаунтов. Проверь, как сбросить именно свои тестовые данные, не затрагивая коллег. Разделяй данные разных команд. Запиши версию сборки и окружение для воспроизводимости проверки.','engineering','team','https://github.com/avito-tech/playbook',true),
 ('qa-test-design','qa','Тест-дизайн первого сценария','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Опиши предусловия, действие и ожидаемый результат. Проверь основной путь, границы, неверный ввод и отсутствие доступа. Отдельно проверь повтор запроса и сбой зависимости. Свяжи проверки с пользовательской задачей, а не с внутренней реализацией.','quality','team','https://github.com/avito-tech/playbook',true),
 ('qa-bug-report','qa','Как оформить баг-репорт','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Укажи окружение, шаги воспроизведения, фактический и ожидаемый результат. Добавь безопасный пример и requestId, если он есть. Не публикуй токены и пароли в приложениях. Проверь повторяемость и обсуди приоритет с владельцем сервиса.','quality','team','https://github.com/avito-tech/playbook',true),
 ('qa-first-task','qa','Первая задача QA','ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Пройди основной путь стажёра: вход, вопрос, подтверждение заявки и обновление чек-листа. Затем проверь другой аккаунт и другую команду. Данные и закрытые инструкции не должны смешиваться. Зафиксируй результат и обсуди находки с бадди.','quality','team','https://github.com/avito-tech/playbook',true) ON CONFLICT DO NOTHING;

UPDATE knowledge_documents SET title='Первые недели в команде',content='ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

В первую неделю познакомься с коллегами, договорись о встречах с бадди и получи доступы. Во вторую — изучи документы команды и возьми первую учебную задачу. В конце месяца собери вопросы и обсуди обратную связь. Чек-лист в профиле помогает увидеть следующий шаг.',updated_at=now() WHERE id='welcome';

UPDATE knowledge_documents SET title='Доступы Backend: Stash',content='ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Репозитории учебной Backend-команды находятся в Stash. Спроси помощника о доступе к Stash и подтверди заявку. После получения номера согласуй её с Алексом. Не отправляй настоящий SSH-ключ в чат; добавление ключа выполняется в самом инструменте после выдачи доступа.',updated_at=now() WHERE id='access-backend';

UPDATE knowledge_documents SET title='Принципы разработки Backend',content='ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Изучи публичный AvitoTech Playbook: тестируемость изменений, обработку сбоев и ясные границы сервисов. В учебной команде HTTP-контракт согласуют до интеграции, изменения проверяют тестами и code review. Конкретные договорённости этой демо-команды вымышлены.',updated_at=now() WHERE id='engineering';

UPDATE knowledge_documents SET title='Рабочие договорённости Analytics',content='ДЕМО-ИНСТРУКЦИЯ. Не является регламентом Авито.

Это командный демо-материал Analytics. Перед исследованием согласуй вопрос, период и единицу наблюдения. Договорённости по метрикам фиксируются в каталоге, а результат обсуждается с бадди Сашей. Стажёрам других команд этот документ не выдаётся.',updated_at=now() WHERE id='analytics-private';

UPDATE corporate_tools SET team_id=NULL,document_id='access-company' WHERE id IN ('confluence','vpn');

UPDATE corporate_tools SET description='Репозитории и code review Backend-команды' WHERE id='stash';

INSERT INTO corporate_tools(id,name,description,team_id,document_id) VALUES
 ('figma','Figma','Макеты и прототипы интерфейсов','frontend','frontend-access'),
 ('storybook','Storybook','Каталог UI-компонентов','frontend','frontend-access'),
 ('metabase','Metabase','Учебные отчёты и дашборды','analytics','analytics-access'),
 ('jupyter','Jupyter','Исследования на синтетических данных','analytics','analytics-access'),
 ('testrail','TestRail','Учебные тест-кейсы','qa','qa-access'),
 ('test-environment','Тестовый стенд','Демо-окружение для проверки сценариев','qa','qa-access') ON CONFLICT DO NOTHING;

UPDATE checklist_tasks SET team_id=NULL,document_id='access-company' WHERE id='access-tools';

INSERT INTO checklist_tasks(id,team_id,title,description,period,sort_order,document_id,source_url) VALUES
 ('read-welcome',NULL,'Прочитать гид новичка','Познакомься с этапами первых недель и сохрани вопросы для бадди.','1 неделя',5,'welcome','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('security-basics',NULL,'Изучить правила безопасности','Узнай, как работать с тестовыми данными и где хранить секреты.','1 неделя',35,'security-basics','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('learning-plan',NULL,'Составить план обучения','Выбери с бадди две технические темы и ожидаемые результаты.','2 неделя',210,'learning-plan','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('first-feedback',NULL,'Получить первую обратную связь','Обсуди первые результаты и договорись о следующем шаге.','1 месяц',300,'feedback','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('meet-community',NULL,'Познакомиться с сообществом стажёров','Найди общий канал и узнай, где можно задавать вопросы.','1 месяц',310,'communications','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('backend-local','backend','Запустить Go-сервис локально','Проверь зависимости, healthcheck и тесты.','1 неделя',100,'backend-local','https://github.com/avito-tech/playbook'),
 ('backend-first-task','backend','Взять первую Backend-задачу','Согласуй ожидаемое поведение и отрицательный сценарий.','2 неделя',110,'backend-first-task','https://github.com/avito-tech/playbook'),
 ('backend-review','backend','Пройти первое code review','Подготовь описание изменения и способ проверки.','2 неделя',120,'backend-review','https://github.com/avito-tech/playbook'),
 ('backend-observability','backend','Разобраться с логами и метриками','Проследи один запрос и обсуди обработку ошибок.','1 месяц',130,'backend-observability','https://github.com/avito-tech/playbook'),
 ('frontend-access','frontend','Получить доступ к Figma и Storybook','Отправь заявку и проверь фактический доступ.','1 неделя',100,'frontend-access','https://github.com/avito-tech/playbook'),
 ('frontend-local','frontend','Запустить веб-приложение','Проверь dev-сервер и production-сборку.','1 неделя',110,'frontend-local','https://github.com/avito-tech/playbook'),
 ('frontend-components','frontend','Изучить компоненты команды','Посмотри состояния и правила доступности.','1 неделя',120,'frontend-components','https://github.com/avito-tech/playbook'),
 ('frontend-first-task','frontend','Взять первую Frontend-задачу','Согласуй макет, состояния и мобильное поведение.','2 неделя',130,'frontend-first-task','https://github.com/avito-tech/playbook'),
 ('frontend-review','frontend','Пройти review интерфейса','Проверь длинный текст, клавиатуру и обработку ошибок.','2 неделя',140,'frontend-review','https://github.com/avito-tech/playbook'),
 ('frontend-quality','frontend','Проверить свой экран на узком устройстве','Убедись, что основной путь доступен на мобильном экране.','1 месяц',150,'frontend-review','https://github.com/avito-tech/playbook'),
 ('analytics-access','analytics','Получить доступ к Metabase и Jupyter','Подтверди заявку и уточни согласование у бадди.','1 неделя',100,'analytics-access','https://github.com/avito-tech/playbook'),
 ('analytics-metrics','analytics','Изучить словарь метрик','Разбери определения, периоды и владельцев метрик.','1 неделя',110,'analytics-metrics','https://github.com/avito-tech/playbook'),
 ('analytics-notebook','analytics','Подготовить первый notebook','Опиши синтетический датасет и воспроизводимый расчёт.','2 неделя',120,'analytics-notebook','https://github.com/avito-tech/playbook'),
 ('analytics-first-task','analytics','Взять первую аналитическую задачу','Согласуй вопрос, результат и ограничения данных.','2 неделя',130,'analytics-first-task','https://github.com/avito-tech/playbook'),
 ('analytics-review','analytics','Обсудить результат исследования','Покажи расчёт и ограничения, запиши вопросы.','2 неделя',140,'analytics-first-task','https://github.com/avito-tech/playbook'),
 ('analytics-quality','analytics','Проверить воспроизводимость отчёта','Попроси коллегу повторить расчёт по описанию.','1 месяц',150,'analytics-notebook','https://github.com/avito-tech/playbook'),
 ('qa-access','qa','Получить доступ к TestRail и стенду','Отправь заявку на нужный инструмент и проверь доступ.','1 неделя',100,'qa-access','https://github.com/avito-tech/playbook'),
 ('qa-local','qa','Подготовить тестовое окружение','Проверь адрес стенда и тестовые аккаунты.','1 неделя',110,'qa-local','https://github.com/avito-tech/playbook'),
 ('qa-test-design','qa','Составить проверки первого сценария','Добавь основной путь, границы, ошибки и права.','1 неделя',120,'qa-test-design','https://github.com/avito-tech/playbook'),
 ('qa-bug-report','qa','Оформить учебный баг-репорт','Запиши шаги, окружение и ожидаемый результат.','2 неделя',130,'qa-bug-report','https://github.com/avito-tech/playbook'),
 ('qa-first-task','qa','Протестировать путь стажёра','Проверь вход, ответ, заявку и чек-лист.','2 неделя',140,'qa-first-task','https://github.com/avito-tech/playbook'),
 ('qa-isolation','qa','Проверить изоляцию разных команд','Сравни два аккаунта и убедись, что данные не смешиваются.','1 месяц',150,'qa-first-task','https://github.com/avito-tech/playbook') ON CONFLICT DO NOTHING;

UPDATE checklist_tasks SET sort_order=200 WHERE id='team-docs';
