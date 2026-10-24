.PHONY: lint
lint:
	go tool golangci-lint run ./...

.PHONY: fmt
fmt:
	go tool golangci-lint fmt

.PHONY: test
test:
	gotestsum --format dots-v2 ./...

.PHONY: bench
bench:
	go test -bench=. -benchmem -run='^$$' ./...
