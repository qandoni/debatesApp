include .env
export

export PROJECT_ROOT=${shell pwd}

minio-up:
	@docker compose up -d minio
minio-down:
	@docker compose down minio

env-up:
	@docker compose up -d debatesApp-postgres
env-down:
	@docker compose down debatesApp-postgres
env-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность потери данных. [y/N]: " ans; \
	if [ "$$ans" = "y" ]; then \
		docker compose down debatesApp-postgres port-forwarder && \
		sudo rm -rf ${PROJECT_ROOT}/out/pgdata && \
		echo "Файлы окружения очищены"; \
	else \
		echo "Очистка окружения отменена"; \
	fi

env-port-forward:
	@docker compose up -d port-forwarder
env-port-close:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует необходимый параметр seq. Пример: make migrate-create seq=init"; \
		exit 1;\
	fi; \
	docker compose run --rm debatesApp-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"
migrate-up:
	@make migrate-action action=up
migrate-down:
	@make migrate-action action=down
migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует необходимый параметр action. Пример: make migrate-action action=up"; \
		exit 1;\
	fi; \
	docker compose run --rm debatesApp-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@debatesApp-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		"$(action)"

debatesApp-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	sudo go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/debates_app/main.go

minio-test-run:
	sudo go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/test_minio/main.go

ps:
	@docker compose ps