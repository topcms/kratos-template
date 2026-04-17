VERSION=$(shell git describe --tags --always 2>/dev/null || echo dev)

.PHONY: init
init:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/wire/cmd/wire@latest

.PHONY: config
# generate internal proto
config:
	cd internal/conf && buf generate

.PHONY: api
api:
	cd proto && buf generate --template ../buf.gen.yaml

.PHONY: wire
wire:
	cd cmd/server && go run github.com/google/wire/cmd/wire

# gorm.io/gen：需 MySQL 可连，DSN 见 configs/config.yaml（或环境变量 GEN_DSN）
.PHONY: gen-all gen-tables
gen-all:
	go run ./cmd/gen -all

# 示例: make gen-tables TABLES=account,admin,user
gen-tables:
	go run ./cmd/gen -tables $(TABLES)

.PHONY: build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./cmd/server

.PHONY: generate
generate:
	go generate ./...
	go mod tidy

.PHONY: all
all: api generate wire

.PHONY: help
help:
	@echo "Targets: init, api, wire, generate, build, gen-all, gen-tables, all"
	@echo "api: run buf generate in ./proto (buf.yaml/buf.lock in proto/, template at ../buf.gen.yaml)"
	@echo "gen-all: gorm gen 生成当前库全部表 (go run ./cmd/gen -all)"
	@echo "gen-tables: gorm gen 指定表，需 TABLES=account,admin,user"

.DEFAULT_GOAL := help
