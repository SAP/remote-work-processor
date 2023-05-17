#!/usr/bin/make -f

OUT_EXEC ?= remote-work-processor
MAIN = ./cmd/remote-work-processor/main.go
PROTO_DIR = build/proto

build: fmt vet
	$(CURDIR)/scripts/build.sh "$(MAIN)" $(OUT_EXEC)

fmt:
	$(CURDIR)/scripts/assertgofmt.sh

vet:
	$(CURDIR)/scripts/assertgovet.sh

proto:
	$(MAKE) -C $(PROTO_DIR)

proto-go:
	$(MAKE) -C $(PROTO_DIR) go-build

proto-go/clean:
	$(MAKE) -C $(PROTO_DIR) go-clean

proto-clean:
	$(MAKE) -C $(PROTO_DIR) clean

clean: proto-clean
	rm $(OUT_EXEC)
	go clean -x

.PHONY: all build fmt vet proto proto-go proto-clean clean