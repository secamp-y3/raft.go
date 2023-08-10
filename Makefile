.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: client
console:
	@docker compose exec client go run main.go
