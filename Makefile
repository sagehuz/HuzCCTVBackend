GO ?= go
PKG := ./...
OUT_DIR := dist
PUBLIC_SRC := public
PUBLIC_DST := cmd/huzbackend/public

.PHONY: build-linux build-macos build-windows build-all sync-public check-public test vet fmt clean

build-linux: sync-public
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-linux-amd64 ./cmd/huzbackend
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-linux-arm64 ./cmd/huzbackend

build-macos: sync-public
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-darwin-amd64 ./cmd/huzbackend
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-darwin-arm64 ./cmd/huzbackend

build-windows: sync-public
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend.exe ./cmd/huzbackend

build-all: build-linux build-macos build-windows

# ------------------------------------------------------------------
# Frontend sync
# public/ is the single source of truth for the web frontend.
# cmd/huzbackend/public/ is the copy embedded into the binary via
# go:embed; it is generated and must never be edited by hand.
# ------------------------------------------------------------------
sync-public:
	@rm -rf $(PUBLIC_DST)
	@mkdir -p $(PUBLIC_DST)
	@cp -R $(PUBLIC_SRC)/. $(PUBLIC_DST)/
	@echo "Synced $(PUBLIC_SRC) -> $(PUBLIC_DST)"

check-public:
	@diff -rq $(PUBLIC_SRC) $(PUBLIC_DST) >/dev/null 2>&1 \
		|| (echo "ERROR: frontend out of sync - run 'make sync-public'" && exit 1)

vet:
	$(GO) vet ./...

test: check-public
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf $(OUT_DIR)
