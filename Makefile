.PHONY: build test lint run provider-validate security

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

security:
	GOTOOLCHAIN=auto go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet -exclude=G304 ./...
	GOTOOLCHAIN=auto go run golang.org/x/vuln/cmd/govulncheck@latest ./...
