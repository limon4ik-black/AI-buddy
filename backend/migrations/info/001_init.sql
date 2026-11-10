CREATE TABLE teams (id text PRIMARY KEY, name text NOT NULL, description text NOT NULL, buddy_id text NOT NULL);
CREATE TABLE people (id text PRIMARY KEY, full_name text NOT NULL, position text NOT NULL);
CREATE TABLE team_members (
 team_id text NOT NULL REFERENCES teams(id), person_id text NOT NULL REFERENCES people(id),
 role text NOT NULL CHECK(role IN ('lead','buddy','engineer')), PRIMARY KEY(team_id,person_id)
);
ALTER TABLE teams ADD CONSTRAINT team_buddy_membership FOREIGN KEY(id,buddy_id) REFERENCES team_members(team_id,person_id) DEFERRABLE INITIALLY DEFERRED;
CREATE VIEW team_directory AS SELECT t.*, p.full_name AS buddy FROM teams t JOIN people p ON p.id=t.buddy_id;
CREATE TABLE technical_services (
 id text PRIMARY KEY, name text NOT NULL, description text NOT NULL,
 owner_team_id text NOT NULL REFERENCES teams(id), technical_owner_id text NOT NULL,
 lifecycle text NOT NULL CHECK(lifecycle IN ('development','active','deprecated')),
 repository_url text NOT NULL, documentation_url text NOT NULL,
 updated_at timestamptz NOT NULL DEFAULT now(),
 FOREIGN KEY(owner_team_id,technical_owner_id) REFERENCES team_members(team_id,person_id)
);
CREATE TABLE profiles (
 user_id text PRIMARY KEY, full_name text NOT NULL DEFAULT 'Стажёр',
 team_id text NOT NULL DEFAULT 'backend' REFERENCES teams(id), position text NOT NULL DEFAULT 'Backend-стажёр',
 start_date date NOT NULL DEFAULT current_date
);
CREATE TABLE knowledge_documents (
 id text PRIMARY KEY, team_id text REFERENCES teams(id), title text NOT NULL, content text NOT NULL,
 category text NOT NULL, visibility text NOT NULL CHECK(visibility IN ('company','team')),
 source_url text NOT NULL, synthetic boolean NOT NULL DEFAULT true,
 updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 CHECK ((visibility='company' AND team_id IS NULL) OR (visibility='team' AND team_id IS NOT NULL))
);
CREATE TABLE corporate_tools (id text PRIMARY KEY, name text NOT NULL, description text NOT NULL, team_id text REFERENCES teams(id), document_id text NOT NULL REFERENCES knowledge_documents(id));
CREATE TABLE checklist_tasks (
 id text PRIMARY KEY, team_id text REFERENCES teams(id), title text NOT NULL, description text NOT NULL,
 period text NOT NULL, sort_order integer NOT NULL, document_id text REFERENCES knowledge_documents(id), source_url text NOT NULL
);
CREATE TABLE user_task_progress (
 user_id text NOT NULL REFERENCES profiles(user_id), task_id text NOT NULL REFERENCES checklist_tasks(id),
 status text NOT NULL CHECK(status IN ('pending','in_progress','completed')), completed_at timestamptz,
 PRIMARY KEY(user_id,task_id), CHECK ((status='completed')=(completed_at IS NOT NULL))
);
CREATE TABLE access_requests (
 id text PRIMARY KEY, user_id text NOT NULL REFERENCES profiles(user_id), system text NOT NULL,
 status text NOT NULL CHECK(status='submitted'), external_id text NOT NULL, idempotency_key text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now(), UNIQUE(user_id,idempotency_key)
);
INSERT INTO teams VALUES
 ('backend','Backend · Demo','Синтетическая команда серверной разработки. Демо-окружение AI Buddy.','alex'),
 ('analytics','Analytics · Demo','Синтетическая команда аналитики.','sasha');
INSERT INTO people VALUES
 ('alex','Алекс','Go-разработчик · Demo'),('maria','Мария','Тимлид · Demo'),
 ('ivan','Иван','Backend-разработчик · Demo'),('sasha','Саша','Аналитик · Demo'),('daria','Дарья','Тимлид аналитики · Demo');
INSERT INTO team_members VALUES
 ('backend','alex','buddy'),('backend','maria','lead'),('backend','ivan','engineer'),
 ('analytics','sasha','buddy'),('analytics','daria','lead');
INSERT INTO technical_services(id,name,description,owner_team_id,technical_owner_id,lifecycle,repository_url,documentation_url) VALUES
 ('intern-profile','Intern Profile API','Демо-сервис профилей стажёров и прогресса адаптации.','backend','alex','active','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('knowledge-api','Knowledge API','Демо-каталог команд, технических сервисов и инструкций.','backend','maria','development','https://github.com/limon4ik-black/AI-buddy','https://github.com/limon4ik-black/AI-buddy'),
 ('onboarding-analytics','Onboarding Analytics','Синтетический сервис аналитики адаптации.','analytics','sasha','development','','');
INSERT INTO knowledge_documents(id,team_id,title,content,category,visibility,source_url) VALUES
 ('welcome',NULL,'Первые недели и работа с наставником','ДЕМО-ИНСТРУКЦИЯ. В первую неделю познакомься с командой, попроси наставника добавить тебя в ключевые чаты, обсуди цели и регулярные 1:1. Во вторую неделю изучи документы команды. Эти направления взяты из публичного чек-листа; организация демо-команды вымышлена.','onboarding','company','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('access-backend','backend','Доступы к инструментам Backend','ДЕМО-ИНСТРУКЦИЯ, не регламент Авито. Для Stash, Confluence или VPN выбери инструмент и подтверди отправку заявки. Mock-сервис вернёт номер и статус submitted. Это только регистрация заявки, доступ ещё не выдан; уточни согласование у демо-наставника.','access','team','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('engineering','backend','Принципы разработки и первые задачи','ДЕМО-МАТЕРИАЛ. Изучи публичный AvitoTech Playbook: инженерные принципы, автоматическое тестирование и поведение при сбоях. Первую учебную задачу и критерии готовности обсуди с наставником. Процесс нашей демо-команды синтетический.','engineering','team','https://github.com/avito-tech/playbook/blob/master/development-principles.md'),
 ('analytics-private','analytics','Закрытая инструкция Analytics','Тестовый документ другой команды. Стажёру Backend недоступен.','engineering','team','https://example.invalid/demo');
INSERT INTO corporate_tools VALUES
 ('stash','Stash','Репозитории демо-команды','backend','access-backend'),
 ('confluence','Confluence','Документация и пространство команды','backend','access-backend'),
 ('vpn','VPN','Подключение к демо-инфраструктуре','backend','access-backend');
INSERT INTO checklist_tasks VALUES
 ('meet-team',NULL,'Познакомиться с командой','Обсуди с наставником состав команды и ключевые чаты.','1 неделя',10,'welcome','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('buddy',NULL,'Договориться о встречах с наставником','Обсуди цели адаптации и регулярные 1:1.','1 неделя',20,'welcome','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('access-tools','backend','Запросить доступ к пространству команды','Можно оформить демо-заявку. Отметь задачу выполненной после проверки фактического доступа.','1 неделя',30,'access-backend','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('stash-keys','backend','Добавить ключи в Stash','Учебный этап по мотивам Backend-чек-листа; настоящие ключи здесь не вводятся.','1 неделя',40,'access-backend','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit'),
 ('principles','backend','Изучить принципы разработки','Прочитай публичный Playbook и обсуди вопросы с наставником.','1 неделя',50,'engineering','https://github.com/avito-tech/playbook/blob/master/development-principles.md'),
 ('team-docs',NULL,'Изучить документы своей команды','Запиши вопросы по рабочим процессам для встречи с наставником.','2 неделя',60,'welcome','https://docs.google.com/spreadsheets/d/1qKFqOzfSZ2jEz57jBSmq7rukDXi_tRfWohgemViceV0/edit');
