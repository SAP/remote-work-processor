#!/usr/bin/make -f

GO_OUTPUT := ./generated/
PROTO_DIR = ../../api/proto

all: go-build java-build

go-build: go-clean
	mkdir -p $(GO_OUTPUT)
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(GO_OUTPUT) \
		--go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUTPUT) \
		--go-grpc_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

go-clean:
	rm -rf $(GO_OUTPUT)

clean: go-clean

.PHONY: all go-build go-clean clean