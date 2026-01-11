SHELL := bash -e

GO := CGO_ENABLED=0 go

unit-test:
	$(GO) test ./...
.PHONY: unit-test

tidy:
	$(GO) mod tidy
.PHONY: tidy

lint: tidy
	$(GO) vet ./...
	$(GO) fmt ./...
.PHONY: lint
