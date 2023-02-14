.PHONY: docs
docs:
	swag init -g ./cmd/server/server.go

.PHONY: build
build:
	go build -o output/tks-api ./cmd/server/server.go

.PHONY: run
run:
	output/tks-api

.PHONY: test
test:
	go test -v ./...

