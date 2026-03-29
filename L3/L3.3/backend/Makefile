.PHONY: help migrate-up migrate-down migrate-create migrate-status docker-up docker-down

GOOSE_CMD = goose -dir internal/migrations postgres "postgres://postgres:postgres@localhost:5432/comment?sslmode=disable"

help:
	@echo "Доступные команды:"
	@echo "  make migrate-up     - применить все миграции"
	@echo "  make migrate-down   - откатить последнюю миграцию"
	@echo "  make migrate-create - создать новую миграцию (используйте NAME=имя)"
	@echo "  make migrate-status - показать статус миграций"
	@echo "  make docker-up      - запустить все контейнеры"
	@echo "  make docker-down    - остановить все контейнеры"
	@echo "  make init           - инициализировать проект (миграции + docker)"

migrate-up:
	$(GOOSE_CMD) up

migrate-down:
	$(GOOSE_CMD) down

migrate-create:
ifndef NAME
	@echo "Укажите NAME, например: make migrate-create NAME=add_users_table"
else
	$(GOOSE_CMD) create $(NAME) sql
endif

migrate-status:
	$(GOOSE_CMD) status

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-clean:
	docker-compose down -v

init: docker-up migrate-up
	@echo "Проект инициализирован!"