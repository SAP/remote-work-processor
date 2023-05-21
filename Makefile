#!/usr/bin/make -f

OUT_DIR ?= dist
OUT_NAME ?= remote-work-processor
OUT_EXEC = $(OUT_DIR)/$(OUT_NAME)
MAIN = ./cmd/remote-work-processor/main.go
PROTO_DIR = build/proto
DOCKER ?= docker
REGISTRY ?= ghcr.io/sap
IMAGE_NAME ?= remote-work-processor
VERSION ?= dev
IMAGE_TAG = $(REGISTRY)/$(IMAGE_NAME):$(VERSION)

build: fmt vet
	$(CURDIR)/scripts/build.sh "$(MAIN)" $(OUT_EXEC)

image: build
	$(DOCKER) build \
		--no-cache \
		--tag $(IMAGE_TAG) \
		--build-arg BIN_FILE=$(OUT_EXEC) \
		.

test:
	go test ./...

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