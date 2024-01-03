.PHONY: lint
lint:
	gocritic check -enableAll ./...

.PHONY: test
test:
	gotestsum --format dots-v2 ./...
