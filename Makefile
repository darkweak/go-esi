.PHONY: lint run-caddy run-server run-traefik

lint: ## Run golangci-lint to ensure the code quality
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run

run-caddy: ## Build and run caddy binary
	cd middleware/caddy && $(MAKE) build && $(MAKE) run

run-server: ## Run server main.go
	go run middleware/server/main.go

run-traefik: ## Run server main.go
	cd middleware/traefik && $(MAKE) build && $(MAKE) run
