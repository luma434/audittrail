dev:
	go run ./cmd/audittrail

test:
	go test ./...

build:
	go build -o bin/audittrail ./cmd/audittrail
