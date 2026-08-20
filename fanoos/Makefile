DB_URL=postgres://postgres:postgres@localhost:5432/fanoos?sslmode=disable
MIGRATION_DIR=scripts/migrations

new:
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $(name)

migrate-up:
	migrate -database "$(DB_URL)" -path $(MIGRATION_DIR) up

migrate-down:
	migrate -database "$(DB_URL)" -path $(MIGRATION_DIR) down 1

migrate-reset:
	migrate -database "$(DB_URL)" -path $(MIGRATION_DIR) down

migrate-force:
	migrate -database "$(DB_URL)" -path $(MIGRATION_DIR) force $(version)

migrate-status:
	migrate -database "$(DB_URL)" -path $(MIGRATION_DIR) version

help:
	@echo "Makefile commands:"
	@echo "  make new name=create_users_table   # Create a new migration"
	@echo "  make migrate-up                    # Apply all migrations"
	@echo "  make migrate-down                  # Rollback last migration"
	@echo "  make migrate-reset                 # Rollback all migrations"
	@echo "  make migrate-force version=1       # Force migration version"
	@echo "  make migrate-status                # Check migration status"