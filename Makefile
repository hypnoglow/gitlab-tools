BIN_DIR=$(shell pwd)/bin

.PHONY: build
build:
	cd ./cmd/gitlab-runner-janitor && \
	go build -o $(BIN_DIR)/gitlab-runner-janitor \
		-trimpath
