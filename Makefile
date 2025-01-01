DB_NAME ?= users
DB_USER ?= postgres
MIGRATION_SCRIPT ?= cmd/db/main.go

run: db-init
	@echo "Starting the application..."
	@go run cmd/main/main.go

db-init:
	@echo "Checking if database '$(DB_NAME)' exists..."
	@if ! psql -U $(DB_USER) -lqt | cut -d \| -f 1 | grep -qw $(DB_NAME); then \
		echo "Database '$(DB_NAME)' does not exist. Creating it..."; \
		psql -U $(DB_USER) -c "CREATE DATABASE $(DB_NAME);"; \
	else \
		echo "Database '$(DB_NAME)' already exists. Skipping creation."; \
	fi

	@echo "Applying migrations..."
	@if go run $(MIGRATION_SCRIPT); then \
		echo "Migrations applied successfully."; \
	else \
		echo "Migrations failed. Check the logs for details."; \
		exit 1; \
	fi

check-deps:
	@echo "Checking dependencies..."
	@command -v psql >/dev/null 2>&1 || { echo "psql is required but not installed. Aborting."; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "go is required but not installed. Aborting."; exit 1; }
	@if [ -z "$$EMAIL_PASSWORD" ]; then \
		echo "EMAIL_PASSWORD is not set. Aborting."; \
		exit 1; \
	fi
	@echo "All dependencies are installed."

clean-db:
	@echo "Dropping database '$(DB_NAME)'..."
	@psql -U $(DB_USER) -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "Database '$(DB_NAME)' dropped."

all: check-deps db-init run

protoc-auth:
	protoc --go_out=. --go-grpc_out=. proto/auth.proto

protoc-password:
	protoc --go_out=. --go-grpc_out=. proto/password.proto

.PHONY: run db-init check-deps clean-db all protoc-auth protoc-password