include .env
export

export PROJECT_ROOT=${shell pwd}

env-up:
	@docker compose up -d debatesApp_postgres
env-down:
	@docker compose down debatesApp_postgres

env-port-forward:
	@docker compose up -d port-forwarder
env-port-close:
	@docker compose down port-forwarder

ps:
	@docker compose ps