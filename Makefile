# grux-poc-api Makefile
# Common developer workflow for the Go HTTP service defined in PLAN-01.md.

GO            ?= go
BINARY        ?= bin/api
IMAGE         ?= grux-poc-api:latest
PKG           ?= ./...
MAIN_PKG      := ./cmd/api

.PHONY: build test run docker-build docker-run docker-clean tidy fmt vet clean help

## help: list available targets
help:
	@echo "Targets:"
	@echo "  build         - compile to $(BINARY)"
	@echo "  test          - run all tests verbosely"
	@echo "  run           - run the API locally on PORT (default 8080)"
	@echo "  tidy          - run go mod tidy"
	@echo "  fmt           - gofmt -w across the module"
	@echo "  vet           - go vet $(PKG)"
	@echo "  docker-build  - build the production image ($(IMAGE))"
	@echo "  docker-run    - run the image locally on :8080"
	@echo "  docker-clean  - remove the local image"
	@echo "  clean         - remove local build artifacts"

## build: compile the API binary into ./bin
build:
	@mkdir -p bin
	$(GO) build -trimpath -ldflags="-s -w" -o $(BINARY) $(MAIN_PKG)

## test: run the full test suite, verbose
test:
	$(GO) test $(PKG) -v

## run: start the API locally (override PORT via env)
run:
	PORT=$${PORT:-8080} $(GO) run $(MAIN_PKG)

## tidy: resolve and pin module dependencies
tidy:
	$(GO) mod tidy

## fmt: format the source tree
fmt:
	$(GO) fmt $(PKG)

## vet: run go vet for suspicious constructs
vet:
	$(GO) vet $(PKG)

## docker-build: build the container image
docker-build:
	docker build -t $(IMAGE) .

## docker-run: run the container, mapping 8080:8080
docker-run:
	docker run --rm -p 8080:8080 -e PORT=8080 $(IMAGE)

## docker-clean: remove the local image
docker-clean:
	docker rmi -f $(IMAGE) 2>/dev/null || true

## clean: remove local build artifacts
clean:
	rm -rf bin
