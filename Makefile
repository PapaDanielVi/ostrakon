.PHONY: build test lint fmt vet all mock

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
MOCKGEN := go tool mockgen
MOCKS_DIR := pkg/mocks

build:
	go build -ldflags "-s -w -X github.com/PapaDanielVi/ostrakon/cmd/ostrakon/commands.Version=$(VERSION)" -o ostrakon ./cmd/ostrakon

test:
	go test ./... -v -race

mock:
	mkdir -p $(MOCKS_DIR)
	$(MOCKGEN) -package=mocks -destination=$(MOCKS_DIR)/mock_keyring.go github.com/PapaDanielVi/ostrakon/pkg/keyring Keyring
	$(MOCKGEN) -package=mocks -destination=$(MOCKS_DIR)/mock_vault.go github.com/PapaDanielVi/ostrakon/pkg/vault Provider

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

vet:
	go vet ./...

all: fmt vet lint test build
