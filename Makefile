run: db-init
	go run cmd/main/main.go

db-init:
	@echo "Checking if database exists..."
	@if ! psql -U postgres -lqt | cut -d \| -f 1 | grep -qw users; then \
		echo "Database 'users' does not exist. Creating it..."; \
		psql -U postgres -c "CREATE DATABASE users;"; \
	else \
		echo "Database 'users' already exists. Skipping creation."; \
	fi
	go run cmd/db/main.go