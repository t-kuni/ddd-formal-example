.PHONY: up down reset verify test integration logs
up:
	docker compose up --build -d

down:
	docker compose down

reset:
	docker compose down -v
	docker compose up --build -d

verify:
	docker compose --profile test run --build --rm verify

test:
	go test ./...

integration:
	docker compose --profile test run --build --rm integration

logs:
	docker compose logs -f gateway lambda event-source-mapping
