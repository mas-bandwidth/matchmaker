# Network Next Makefile

MODULE ?= "github.com/networknext/matchmaker"

BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

.PHONY: build
build: dist/matchmaker dist/process

.PHONY: format
format:
	@gofmt -s -w .

.PHONY: clean
clean: ## clean everything
	@rm -rf dist
	@mkdir dist

.PHONY: rebuild
rebuild: clean build ## rebuild everything

dist/%: cmd/%/*.go
	@go build -o $@ $(<D)/*.go

