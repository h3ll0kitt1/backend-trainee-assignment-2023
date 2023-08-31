swagger:
	swag init -g ./cmd/main.go -o .swagger -ot yaml

build:
	docker-compose build app

run:
	docker-compose up app

stop:
	docker-compose down