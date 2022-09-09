.PHONY: build run

build: ## Build caddy binary
	go mod tidy
	go mod download
	xcaddy build --with github.com/darkweak/go-esi@latest=../.. --with github.com/darkweak/go-esi/middleware/caddy=./

run: ## Run caddy with go-esi
	./caddy run