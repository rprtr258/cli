.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	golangci-lint fmt

.PHONY: test
test:
	gotestsum --format dots-v2 ./...
