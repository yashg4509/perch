.PHONY: build test lint run provider-validate

build:
	go build -o bin/perch ./cmd/perch

test:
	go test ./...

lint:
	go vet ./...
	go fmt ./...

run:
	go run ./cmd/perch

provider-validate:
	go test ./internal/providerspec/... -count=1
