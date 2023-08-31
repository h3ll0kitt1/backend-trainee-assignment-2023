swagger:
	swag init -g ./cmd/main.go -o .swagger -ot yaml

run:
	docker-compose up --build app

stop:
	docker-compose down