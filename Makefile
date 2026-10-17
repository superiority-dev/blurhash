SHELL := /bin/bash
.SHELLFLAGS := -o pipefail -ec

GOLANGCI_LINT ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0
GOVULNCHECK   ?= go run golang.org/x/vuln/cmd/govulncheck@v1.9.0

.PHONY: check fmt vet lint test vuln fma

check: fmt vet lint test vuln fma

fmt:
	@dirs=$$(go list -f '{{.Dir}}' ./...); \
	out=$$(gofmt -l $$dirs); if [ -n "$$out" ]; then echo "gofmt:"; echo "$$out"; exit 1; fi

vet:
	go vet ./...

lint:
	$(GOLANGCI_LINT) run ./...

test:
	go test ./...
	go test -race ./...

vuln:
	$(GOVULNCHECK) ./...

fma:
	@dir=$$(mktemp -d); trap 'rm -rf "$$dir"' EXIT; \
	for target in "GOARCH=arm64" "GOARCH=amd64 GOAMD64=v3"; do \
		for pkg in . ./internal/codec ./internal/fdlibm; do \
			env $$target go build -o "$$dir/pkg.a" $$pkg; \
			asm=$$(go tool objdump "$$dir/pkg.a"); \
			fused=$$(grep -E '\bV?FN?M(ADD|SUB)' <<< "$$asm" || true); \
			if [ -n "$$fused" ]; then echo "fused multiply-add in $$pkg with $$target:"; echo "$$fused"; exit 1; fi; \
			echo "fma: none in $$pkg with $$target"; \
		done; \
	done
