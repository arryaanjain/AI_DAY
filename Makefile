API_DIR := api

.PHONY: setup dev api worker test lint migrate-up migrate-down docker-up docker-down
setup:
	cd ui && npm install
	cd admin && npm install
	cd $(API_DIR) && go mod download
dev:
	@echo "Run 'make api', 'cd ui && npm run dev', and 'cd admin && npm run dev' in separate terminals."
api:
	cd $(API_DIR) && go run ./cmd/api
worker:
	cd $(API_DIR) && go run ./cmd/worker
test:
	cd $(API_DIR) && go test ./...
lint:
	cd ui && npm run lint
	cd admin && npm run lint
	cd $(API_DIR) && go vet ./...
migrate-up:
	@echo "Migration runner is introduced with the database phase. SQL migrations are in api/migrations."
migrate-down:
	@echo "Migration runner is introduced with the database phase."
docker-up:
	docker compose up --build -d
docker-down:
	docker compose down
