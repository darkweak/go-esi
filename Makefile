.PHONY: lint run-caddy run-server run-traefik
MIDDLEWARES_LIST=caddy roadrunner server traefik

lint: ## Run golangci-lint to ensure the code quality
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run

run-caddy: ## Build and run caddy binary
	cd middleware/caddy && $(MAKE) build && $(MAKE) run

run-roadrunner: ## Build and run roadrunner
	cd middleware/roadrunner && $(MAKE) build && $(MAKE) run

run-server: ## Run server main.go
	go run middleware/server/main.go

run-traefik: ## Build and run træfik
	cd middleware/traefik && $(MAKE) build && $(MAKE) run

vendor: ## Generate and prepare vendors for each plugin
	go mod tidy && go mod download
	for middleware in $(MIDDLEWARES_LIST) ; do \
	cd middleware/$$middleware && ($(MAKE) build || true) && cd -; done
