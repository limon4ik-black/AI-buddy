#!/usr/bin/env bash
set -euo pipefail
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname postgres \
 -v auth_password="$AUTH_DB_PASSWORD" -v info_password="$INFO_DB_PASSWORD" -v chat_password="$CHAT_DB_PASSWORD" <<'SQL'
CREATE ROLE auth_user LOGIN PASSWORD :'auth_password';
CREATE ROLE info_user LOGIN PASSWORD :'info_password';
CREATE ROLE chat_user LOGIN PASSWORD :'chat_password';
CREATE DATABASE auth_db OWNER auth_user;
CREATE DATABASE info_db OWNER info_user;
CREATE DATABASE chat_db OWNER chat_user;
REVOKE CONNECT ON DATABASE auth_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE info_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE chat_db FROM PUBLIC;
GRANT CONNECT ON DATABASE auth_db TO auth_user;
GRANT CONNECT ON DATABASE info_db TO info_user;
GRANT CONNECT ON DATABASE chat_db TO chat_user;
SQL
