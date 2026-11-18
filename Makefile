# kready 统一命令入口。AGENTS.md 中的命令以此为准。
# 注意：本文件属于门禁，修改需人工批准。

GO ?= go
PKG := ./...
# go.mod 是 Go 版本的唯一来源；CI 与 Makefile 均从此读取，不再手写。
GO_REQUIRED := $(shell awk '$$1=="go"{print $$2}' go.mod)
export GOCACHE ?= $(CURDIR)/.gocache
GO_VERSION := $(shell $(GO) env GOVERSION | sed 's/^go//')

.PHONY: fmt lint test check build clean goversion hooks

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

# 版本由 git tag 驱动，注入 main.version；无 tag 时回退为提交号。
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

build: goversion
	$(GO) build -ldflags "$(LDFLAGS)" -o bin/kready ./cmd/kready

# 启用仓库自带的 commit-msg 钩子（.githooks/），每次 make check 时确保已生效。
hooks:
	@git config core.hooksPath .githooks

check: hooks lint test build

clean:
	rm -rf .gocache
	rm -rf bin
