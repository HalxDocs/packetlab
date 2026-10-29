GO      ?= go
BIN     := bin/packetlab
SCENARIOS := scenarios

.PHONY: all build test test-race cover fmt vet check run traces web-install web-build web-test clean

all: build

build:
	$(GO) build -trimpath -o $(BIN) ./cmd/packetlab

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

cover:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

check: fmt vet test

run: build
	./$(BIN) run $(SCENARIOS)/two-hosts-one-router.json

# Render every scenario into a self-contained trace bundle for the web UI.
traces: build
	@mkdir -p web/public/traces
	@for f in $(SCENARIOS)/*.json; do \
		id=$$(basename $$f .json); \
		./$(BIN) trace $$f > web/public/traces/$$id.json; \
	done
	@ls -1 web/public/traces

web-install:
	cd web && npm install

web-build:
	cd web && npm run build

web-test:
	cd web && npm test -- --run

clean:
	rm -rf bin coverage.out web/dist web/public/traces
