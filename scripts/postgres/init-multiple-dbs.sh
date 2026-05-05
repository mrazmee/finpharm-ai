#!/bin/sh
set -eu

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
  SELECT 'CREATE DATABASE inventory_db'
  WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'inventory_db'
  )\gexec

  SELECT 'CREATE DATABASE transaction_db'
  WHERE NOT EXISTS (
    SELECT FROM pg_database WHERE datname = 'transaction_db'
  )\gexec
EOSQL