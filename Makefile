SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -ec

GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
GOVULNCHECK   ?= go run golang.org/x/vuln/cmd/govulncheck@v1.9.0

.PHONY: check fmt vet lint test vuln

check: fmt vet lint test vuln

fmt:
	@out=$$(gofmt -l $$(go list -f '{{.Dir}}' ./...)); if [ -n "$$out" ]; then echo "gofmt:"; echo "$$out"; exit 1; fi

vet:
	go vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

test:
	go test -race ./...

vuln:
	$(GOVULNCHECK) ./...
