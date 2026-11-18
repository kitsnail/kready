# kready 统一命令入口。AGENTS.md 中的命令以此为准。
# 注意：本文件属于门禁，修改需人工批准。

GO ?= go
PKG := ./...
GO_REQUIRED := 1.27.4
export GOCACHE ?= $(CURDIR)/.gocache
GO_VERSION := $(shell $(GO) env GOVERSION | sed 's/^go//')

.PHONY: fmt lint test check build clean goversion

goversion:
	@if [ "$(GO_VERSION)" != "$(GO_REQUIRED)" ]; then \
		echo "需要 Go $(GO_REQUIRED)，当前为 $(GO_VERSION)"; exit 1; \
	fi

fmt: goversion
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "以下文件需要 gofmt:"; echo "$$out"; \
		gofmt -w $$out; \
	fi

lint: goversion
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "格式不符（请运行 make fmt）:"; echo "$$out"; exit 1; \
	fi
	$(GO) vet $(PKG)

test: goversion
	$(GO) test -race -count=1 $(PKG)

build: goversion
	$(GO) build -o bin/kready ./cmd/kready

check: lint test build

clean:
	rm -rf .gocache
	rm -rf bin
