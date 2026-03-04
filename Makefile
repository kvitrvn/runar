BINARY_NAME=autogest
BINARY_DIR=./bin
CMD_DIR=./cmd/autogest
MODULE=github.com/kvitrvn/runar

.PHONY: all build run test test-legal test-coverage lint clean install help

all: build

## build: Compile le binaire
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=1 go build -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

## run: Lance l'application en mode dev
run:
	CGO_ENABLED=1 go run $(CMD_DIR)/main.go

## test: Lance tous les tests
test:
	CGO_ENABLED=1 go test ./... -v

## test-legal: Lance uniquement les tests des règles légales
test-legal:
	CGO_ENABLED=1 go test ./internal/domain/... ./internal/service/... -v -run "Legal|Immut|Number|Validation"

## test-coverage: Lance les tests avec rapport de couverture
test-coverage:
	CGO_ENABLED=1 go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Lance le linter golangci-lint
lint:
	golangci-lint run ./...

## clean: Supprime les artefacts de build
clean:
	@rm -rf $(BINARY_DIR)
	@rm -f coverage.out coverage.html
	@echo "Cleaned."

## install: Installe le binaire dans GOPATH/bin
install: build
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)

## help: Affiche cette aide
help:
	@echo "Commandes disponibles :"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
