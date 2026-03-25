.PHONY: tidy style lint test build

tidy:
	go mod tidy

style:
	goimports -l -w ./

lint:
	staticcheck ./...

test:
	go test ./...

build:
	go build -o bin/assay ./cmd/assay
