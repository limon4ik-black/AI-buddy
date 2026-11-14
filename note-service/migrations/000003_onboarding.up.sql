CREATE SCHEMA IF NOT EXISTS onboarding;
CREATE TABLE onboarding.knowledge_documents (
 id text PRIMARY KEY, team_id text, title text NOT NULL, content text NOT NULL,
 category text NOT NULL, visibility text NOT NULL CHECK(visibility IN ('company','team')),
 source_url text NOT NULL, synthetic boolean NOT NULL DEFAULT true,
 updated_at timestamptz NOT NULL DEFAULT now(), deleted_at timestamptz,
 CHECK ((visibility='company' AND team_id IS NULL) OR (visibility='team' AND team_id IS NOT NULL))
);
CREATE TABLE onboarding.checklist_tasks (
 id text PRIMARY KEY, team_id text, title text NOT NULL, description text NOT NULL,
 period text NOT NULL, sort_order integer NOT NULL,
 document_id text REFERENCES onboarding.knowledge_documents(id), source_url text NOT NULL
);
CREATE TABLE onboarding.user_task_progress (
 user_id text NOT NULL, task_id text NOT NULL REFERENCES onboarding.checklist_tasks(id),
 status text NOT NULL CHECK(status IN ('pending','in_progress','completed')), completed_at timestamptz,
 PRIMARY KEY(user_id,task_id), CHECK((status='completed')=(completed_at IS NOT NULL))
);
CREATE TABLE onboarding.identity_links (user_id text PRIMARY KEY, source_user_id text NOT NULL);
CREATE TABLE onboarding.imports (name text PRIMARY KEY, completed_at timestamptz NOT NULL DEFAULT now());
