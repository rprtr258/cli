.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: fmt
fmt:
	golangci-lint fmt

.PHONY: test
test:
	gotestsum --format dots-v2 ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem -run='^$$' ./...
