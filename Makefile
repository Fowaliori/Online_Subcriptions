install-swag:
	go install github.com/swaggo/swag/v2/cmd/swag@latest

docs: install-swag
	$(shell go env GOPATH)/bin/swag init -g main.go -d .,./internal/handlers,./internal/api --ot yaml

run:
	docker-compose up --build -d