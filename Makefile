# Matchmaker makefile

.PHONY: build
build: dist/matchmaker dist/transform dist/datacenters dist/combine dist/example dist/average

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
