FROM postgres:14

COPY ../db-init/postgres-init.sql /docker-entrypoint-initdb.d/