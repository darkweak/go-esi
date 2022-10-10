.PHONY: build run

build: ## Build caddy binary
	go mod tidy
	go mod download

run: ## Run caddy with go-esi
	docker-compose up --remove-orphans