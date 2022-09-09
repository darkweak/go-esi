.PHONY: lint run-caddy run-roadrunner

lint: ## Run golangci-lint to ensure the code quality
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint golangci-lint run

run-caddy: ## Build caddy binary
	cd middleware/caddy && $(MAKE) build && $(MAKE) run
