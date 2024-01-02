.PHONY: lint
lint:
	gotestsum --format dots-v2 ./...

.PHONY: check_crosscompile
check_crosscompile:
	./contrib/check_crosscompile.sh
