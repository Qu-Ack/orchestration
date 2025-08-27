APP_BIN=bin/o
APP_CMD=./cmd/
POSTGRES_CONTAINER=my-postgres
POSTGRES_NETWORK=db-network
POSTGRES_PORT=5433

.PHONY: dev prod build run traefik postgres

check-traefik:
	@if docker ps -a --format '{{.Names}}' | grep -q '^traefik$$'; then \
		echo "⚠️  Removing existing traefik container..."; \
		docker stop traefik && docker rm traefik; \
	fi; \
	if [ "$$ENV" = "dev" ]; then \
		cd examples/traefik_init/dev && \
		docker compose up -d; \
	else \
		cd examples/traefik_init && \
		docker compose up -d; \
	fi

check-postgres:
	@if ! docker network ls | grep -q $(POSTGRES_NETWORK); then \
		docker network create $(POSTGRES_NETWORK); \
	fi; \
	if docker ps -a --format '{{.Names}}' | grep -q '^$(POSTGRES_CONTAINER)$$'; then \
		docker stop $(POSTGRES_CONTAINER) && docker rm $(POSTGRES_CONTAINER); \
	fi; \
	docker run -d \
		--name $(POSTGRES_CONTAINER) \
		--network $(POSTGRES_NETWORK) \
		-e POSTGRES_USER=postgres \
		-e POSTGRES_PASSWORD=postgres \
		-e POSTGRES_DB=postgres \
		-v postgres-data:/var/lib/postgresql/data \
		-p $(POSTGRES_PORT):5432 \
		postgres:latest

build:
	@go build -o $(APP_BIN) $(APP_CMD)

dev: 
	@ENV=dev $(MAKE) check-traefik
	@$(MAKE) check-postgres
	@$(MAKE) build
	@sudo ./$(APP_BIN)

prod:
	@if [ -z "$$DO_AUTH_TOKEN" ]; then \
		echo "❌ DO_AUTH_TOKEN is not set. Please obtain a DigitalOcean Personal Access Token for DNS challenge."; \
		exit 1; \
	fi
	@ENV=prod $(MAKE) check-traefik
	@$(MAKE) check-postgres
	@$(MAKE) build
	@sudo ./$(APP_BIN) -env=production

