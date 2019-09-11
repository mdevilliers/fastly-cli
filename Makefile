PACKAGES := $(shell go list ./... | grep -v /vendor/ )
BINDIR   := $(CURDIR)/bin

GO_TEST = go test -covermode=atomic
GO_COVER = go tool cover
GO_BENCH = go test -bench=.

GO_BIN    := $(BINDIR)
PATH     := $(GOBIN):$(PATH)

OS := $(shell uname)
GOLANGCI_LINT_VERSION=1.18.0
ifeq ($(OS),Darwin)
	GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-darwin-amd64.tar.gz
else
	GOLANGCI_LINT_ARCHIVE=golangci-lint-$(GOLANGCI_LINT_VERSION)-linux-amd64.tar.gz
endif


all: test

.PHONY: all

.PHONY: test
test:
ifeq ("$(wildcard $(shell which gocov))","")
	go get github.com/axw/gocov/gocov
endif
	gocov test ${PACKAGES} | gocov report

clean: ## clean up
	rm -rf tmp/

.PHONY: clean

lint: $(GO_BIN)/golangci-lint/golangci-lint ## lint
	$(GO_BIN)/golangci-lint/golangci-lint run

.PHONY: lint

$(GO_BIN)/golangci-lint/golangci-lint:
	curl -OL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_LINT_VERSION)/$(GOLANGCI_LINT_ARCHIVE)
	mkdir -p $(GO_BIN)/golangci-lint/
	tar -xf $(GOLANGCI_LINT_ARCHIVE) --strip-components=1 -C $(GO_BIN)/golangci-lint/
	chmod +x $(GO_BIN)/golangci-lint
	rm -f $(GOLANGCI_LINT_ARCHIVE)

# 'help' parses the Makefile and displays the help text
help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: help
