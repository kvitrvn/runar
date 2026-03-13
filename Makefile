BINARY_NAME=runar
BINARY_DIR=./bin
CMD_DIR=./cmd/runar
MODULE=github.com/kvitrvn/runar
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS=-ldflags "-X main.Version=$(VERSION) -s -w"
GO_GOBIN=$(strip $(or $(GOBIN),$(shell go env GOBIN)))
INSTALL_DIR=$(if $(GO_GOBIN),$(GO_GOBIN),$(shell go env GOPATH)/bin)

# Supprime les warnings C de go-sqlite3 (-Wdiscarded-qualifiers)
export CGO_CFLAGS := -Wno-discarded-qualifiers
export CGO_ENABLED := 1

.PHONY: all build run test test-legal test-coverage lint clean install release help

all: build

## build: Compile le binaire
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BINARY_DIR)
	go build $(BUILD_FLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) $(CMD_DIR)

## release: Compile un binaire optimisé (stripped) pour la plateforme courante
## Note: CGO_ENABLED=1 (sqlite3) requiert gcc pour la cross-compilation.
## Pour Linux AMD64 uniquement depuis Linux: GOOS=linux GOARCH=amd64 make release
release:
	@echo "Building release $(VERSION)..."
	@mkdir -p $(BINARY_DIR)/release
	go build $(BUILD_FLAGS) -o $(BINARY_DIR)/release/$(BINARY_NAME) $(CMD_DIR)
	@ls -lh $(BINARY_DIR)/release/$(BINARY_NAME)
	@echo "Release binary: $(BINARY_DIR)/release/$(BINARY_NAME) (version: $(VERSION))"

## run: Lance l'application en mode dev
run:
	go run $(CMD_DIR)/main.go

## test: Lance tous les tests
test:
	go test ./... -v

## test-legal: Lance uniquement les tests des règles légales
test-legal:
	go test ./internal/domain/... ./internal/service/... -v -run "Legal|Immut|Number|Validation"

## test-coverage: Lance les tests avec rapport de couverture
test-coverage:
	go test ./... -coverprofile=coverage.out
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

## install: Installe le binaire dans GOBIN ou GOPATH/bin
install: build
	@mkdir -p $(INSTALL_DIR)
	install -m 0755 $(BINARY_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Installed: $(INSTALL_DIR)/$(BINARY_NAME)"

## help: Affiche cette aide
help:
	@echo "Commandes disponibles :"
	@sed -n 's/^## //p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
