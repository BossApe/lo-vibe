#!/bin/sh
set -e

# forgejoデータベースが無ければ作成
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" -c "CREATE DATABASE forgejo OWNER $POSTGRES_USER;" || true

# forgejoスキーマが無ければ作成（postgres DB）
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "postgres" -c "CREATE SCHEMA IF NOT EXISTS forgejo;"

# forgejoスキーマが無ければ作成（musuhi DB）
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" -c "CREATE SCHEMA IF NOT EXISTS forgejo;"
# forgejoスキーマが無ければ作成（forgejo DB）
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "forgejo" -c "CREATE SCHEMA IF NOT EXISTS forgejo;"

# musuhi DBにsystem_overviewsテーブルが無ければ作成
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<'EOSQL'
CREATE TABLE IF NOT EXISTS system_overviews (
	id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
	content    TEXT        NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
EOSQL
