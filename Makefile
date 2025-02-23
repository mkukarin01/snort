run:
	go run cmd/shortener/main.go

test:
	go test ./...

coverage:
	go test -cover ./...
