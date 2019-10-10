migrate:
	go run cmd/create-resources/main.go

test:
	go vet ./...
	go test -race -short ./...

build:
	docker build . -t w32blaster.me/bot-weather-watcher