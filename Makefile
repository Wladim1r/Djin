docker-up:
	docker compose up

docker-down:
	docker compose down -v

doc-up-n-c:
	docker compose build --no-cache && docker compose up

swag:
	swag init -g ./cmd/main.go --parseDependency --parseInternal
