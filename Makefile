run:
	go run cmd/shortener/main.go

test:
	go test ./...

coverage:
	go test -cover ./...

mock_storager:
	mockgen -source=internal/storage/storage.go -destination=internal/storage/mock_storage.go -package=storage Storager