GO ?= go

.PHONY: build test fmt vet demo check tidy

build:
	$(GO) build -o bin/herald ./cmd/herald

test:
	$(GO) test ./...

fmt:
	gofmt -w .

vet:
	$(GO) vet ./...

demo:
	$(GO) run ./cmd/herald demo

# The full quality gate: format, vet, test.
check: fmt vet test

tidy:
	$(GO) mod tidy
