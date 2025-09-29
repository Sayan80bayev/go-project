# Environment variables for PostgreSQL (override in .env or environment)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= password
DB_NAME ?= engagement_service
DB_SSLMODE ?= disable

# Database connection string
DB_URL = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Migration directory
MIGRATIONS_DIR = migrations

# Default target
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make migrate-up        # Apply all up migrations"
	@echo "  make migrate-down      # Roll back one migration"
	@echo "  make migrate-status    # Check current migration status"
	@echo "  make migrate-create NAME=name  # Create a new migration file with the given name"
	@echo "  make migrate-force VERSION=version  # Force set migration version (use with caution)"

# Apply all up migrations
.PHONY: migrate-up
migrate-up:
	@echo "Applying migrations..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# Roll back one migration
.PHONY: migrate-down
migrate-down:
	@echo "Rolling back one migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

# Check migration status
.PHONY: migrate-status
migrate-status:
	@echo "Checking migration status..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# Create a new migration file
.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please specify a migration name with NAME=<migration_name>"; \
		exit 1; \
	fi
	@echo "Creating new migration file: $(NAME)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(NAME)

# Force set migration version (use with caution)
.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: Please specify a version with VERSION=<version_number>"; \
		exit 1; \
	fi
	@echo "Forcing migration version to $(VERSION)..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" force $(VERSION)