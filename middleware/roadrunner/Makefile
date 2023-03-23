.PHONY: build log run

build: ## Prepare roadrunner deps
	go mod tidy
	go mod download

log: ## Display container logs
	cd examples && docker-compose -f docker-compose.yml.test logs -f

run: ## Run roadrunner with go-esi
	cd examples && docker-compose -f docker-compose.yml.test up --remove-orphans --build
