all: run

run:
	rm log
	DB_HOST=localhost go run ./src/cmd | tee -a log

migrate:
	@echo "Starting database migrations..."
	@for f in ./migrations/*.sql; do \
		echo "Applying: $$(basename $$f)"; \
		if psql languasia languasia -f "$$f" 2> /tmp/migration-error; then \
			echo "✓ Success: $$(basename $$f)"; \
		else \
			echo "✗ Failed: $$(basename $$f)"; \
			cat /tmp/migration-error; \
			exit 1; \
		fi; \
	done
	@echo "All migrations completed successfully."

