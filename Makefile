migrate:
	go build cmd/bot/setup.go
	./setup
	rm ./setup