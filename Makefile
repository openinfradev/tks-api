.PHONY: docs
docs:
	swag init -g ./cmd/server/main.go --parseDependency --parseInternal -o ./api/swagger

.PHONY: build
build:
	go build -o output/tks-api ./cmd/server/main.go
	go build -o output/tks ./cmd/client/main.go

.PHONY: run
run:
	output/tks-api

.PHONY: test
test:
	go test -v ./...

.PHONY: dev_run
dev_run: 
	swag init -g ./cmd/server/main.go --parseDependency --parseInternal -o ./api/swagger
	go build ./cmd/server/main.go
	./main
