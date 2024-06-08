# Matchmaker makefile

.PHONY: build
build: dist/matchmaker dist/transform dist/datacenters dist/combine dist/example dist/average

.PHONY: format
format:
	@gofmt -s -w .

.PHONY: clean
clean: ## clean everything
	@rm -rf data/players.csv
	@rm -rf dist
	@mkdir dist

.PHONY: rebuild
rebuild: clean build ## rebuild everything

data/players.csv: data/players.zip
	@cd data && unzip -oq players.zip && touch players.csv

dist/%: cmd/%/*.go data/players.csv
	@go build -o $@ $(<D)/*.go
