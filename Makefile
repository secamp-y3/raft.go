.PHONY: up
up:
	docker compose up -d

.PHONY: down
down:
	docker compose down

.PHONY: console
console:
	@docker compose exec console go run main.go
