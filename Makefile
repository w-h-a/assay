.PHONY: tidy style lint test unit-test build

tidy:
	go mod tidy

style:
	goimports -l -w ./

lint:
	staticcheck ./...

test:
	@echo "=== TESTS ==="
	go clean -testcache && INTEGRATION=1 go test -v ./...

unit-test:
	@echo "=== UNIT TESTS ==="
	go clean -testcache && go test -v ./...

build:
	CGO_ENABLED=0 go build -o bin/assay ./cmd/assay
