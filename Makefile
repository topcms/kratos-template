GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always 2>/dev/null || echo dev)

ifeq ($(GOHOSTOS), windows)
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name '*.proto' 2>/dev/null" || true)
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find proto -name '*.proto' 2>/dev/null" || true)
else
	INTERNAL_PROTO_FILES=$(shell find internal -name '*.proto')
	API_PROTO_FILES=$(shell find proto -name '*.proto')
endif

.PHONY: init
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
config:
	protoc --proto_path=./internal \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./internal \
		$(INTERNAL_PROTO_FILES)

.PHONY: api
api:
	protoc --proto_path=./proto \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		$(API_PROTO_FILES)

.PHONY: wire
wire:
	cd cmd/server && go run github.com/google/wire/cmd/wire

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./cmd/server

.PHONY: generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
all: api config generate wire

.PHONY: help
help:
	@echo "Targets: init, api, config, wire, generate, build, all"

.DEFAULT_GOAL := help
