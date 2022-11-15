# Network Next Makefile

MODULE ?= "github.com/networknext/matchmaker"

BUILD_TIME ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
COMMIT_MESSAGE ?= $(shell git log -1 --pretty=%B | tr "\n" " " | tr \' '*')
COMMIT_HASH ?= $(shell git rev-parse --short HEAD) 

matchmaker: matchmaker.go $(shell find . -name '*.go')
	@go build -ldflags "-s -w -X $(MODULE).buildTime=$(BUILD_TIME) -X \"$(MODULE).commitMessage=$(COMMIT_MESSAGE)\" -X $(MODULE).commitHash=$(COMMIT_HASH)" -o $@ $(<D)/*.go

# Format code

.PHONY: format
format:
	@gofmt -s -w .

# Clean and rebuild

.PHONY: clean
clean: ## clean everything
	@rm matchmaker

.PHONY: build
build: matchmaker

.PHONY: rebuild
rebuild: clean build ## rebuild everything
