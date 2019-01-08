GO := go

# environment
export GO111MODULE = on
export CGO_ENABLED = 0

all: test build

test: ## run tests
	$(GO) test -v ./...

build: ## build program
	$(GO) build

# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := help
HELP_PATTERN ?= ^[a-zA-Z_-]+:.*?\#\# .*$$
HELP_MAX_LEN := $(shell \
  grep -E '$(HELP_PATTERN)' $(MAKEFILE_LIST) \
  | awk 'BEGIN {FS = ":"}; {print $$1}' \
  | awk '{ if (length($$0) > max) max = length($$0) } END { print max }' \
)
HELP_DESC_OFFSET ?= $(shell expr $(HELP_MAX_LEN) + 6)

.PHONY: help
help:
	@grep -E '$(HELP_PATTERN)' $(MAKEFILE_LIST) \
		| sort \
		| awk 'BEGIN {FS = ":.*?## "}; \
			{printf "  \033[36m%-$(HELP_DESC_OFFSET)s\033[0m %s\n", $$1, $$2}'
