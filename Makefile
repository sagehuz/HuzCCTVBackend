GO ?= go
PKG := ./...
OUT_DIR := dist

.PHONY: build-linux build-macos build-windows build-all test vet fmt clean

build-linux:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-linux-amd64 ./cmd/huzbackend
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-linux-arm64 ./cmd/huzbackend

build-macos:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-darwin-amd64 ./cmd/huzbackend
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend-darwin-arm64 ./cmd/huzbackend

build-windows:
	@mkdir -p $(OUT_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags="-s -w" -o $(OUT_DIR)/huzbackend.exe ./cmd/huzbackend

build-all: build-linux build-macos build-windows

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

fmt:
	$(GO) fmt ./...

clean:
	rm -rf $(OUT_DIR)
