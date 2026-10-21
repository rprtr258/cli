.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: test
test:
	gotestsum --format dots-v2 ./...
