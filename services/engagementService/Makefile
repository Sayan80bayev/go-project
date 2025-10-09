POSTGRES_HOST ?= localhost
POSTGRES_PORT ?= 5432
POSTGRES_USER ?= postgres
POSTGRES_PASSWORD ?= password
POSTGRES_DB_NAME ?= engagement_db
DB_SSLMODE ?= disable

DB_URL = jdbc:postgresql://$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB_NAME)?sslmode=$(DB_SSLMODE)
MIGRATIONS_DIR = migrations

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make migrate-up          # Apply all pending migrations"
	@echo "  make migrate-down        # Rollback the last migration (Enterprise only or manual)"
	@echo "  make migrate-status      # Show migration status"
	@echo "  make migrate-create NAME=name  # Create a new migration file"
	@echo "  make migrate-force VERSION=version  # Repair schema history to a specific version"

.PHONY: migrate-up
migrate-up:
	flyway -url="$(DB_URL)" -locations=filesystem:$(MIGRATIONS_DIR) migrate

.PHONY: migrate-down
migrate-down:
	@echo "WARNING: Flyway Community Edition does not support undo migrations."
	@echo "To rollback, ensure an undo script (e.g., U1__undo_create_table.sql) exists or use Flyway Enterprise."
	@echo "Running 'flyway undo' (if Enterprise is available)..."
	flyway -url="$(DB_URL)" -locations=filesystem:$(MIGRATIONS_DIR) undo || echo "Undo not supported in Community Edition. Manually apply rollback SQL."

.PHONY: migrate-status
migrate-status:
	flyway -url="$(DB_URL)" -locations=filesystem:$(MIGRATIONS_DIR) info

.PHONY: migrate-create
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "Error: Please specify a migration name with NAME=<migration_name>"; \
		exit 1; \
	fi
	@TIMESTAMP=$$(date +%Y%m%d%H%M%S); \
	FILE=$(MIGRATIONS_DIR)/V$$TIMESTAMP__$(NAME).sql; \
	touch $$FILE; \
	echo "-- Migration: $(NAME)" > $$FILE; \
	echo "-- Add your SQL here" >> $$FILE; \
	echo "Created migration file: $$FILE"

.PHONY: migrate-force
migrate-force:
	@if [ -z "$(VERSION)" ]; then \
		echo "Error: Please specify a version with VERSION=<version_number>"; \
		exit 1; \
	fi
	@echo "Running 'flyway repair' to align schema history and forcing version $(VERSION)..."
	flyway -url="$(DB_URL)" -locations=filesystem:$(MIGRATIONS_DIR) repair
	@echo "Note: Flyway does not directly support forcing a specific version. Use 'repair' and ensure migrations align."