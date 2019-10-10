migrate:
	go build cmd/create-resources/main.go
	./setup
	rm ./setup

test:
	go vet ./...
	go test -race -short ./...

build:
	docker build . -t w32blaster.me/bot-weather-watcher