migrate:
	go build cmd/bot/setup.go
	./setup
	rm ./setup

test:
	go vet ./...
	go test -race -short ./...