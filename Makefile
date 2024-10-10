ifndef GOPATH
	GOPATH := $(shell go env GOPATH)
endif
ifndef GOBIN # derive value from gopath (default to first entry, similar to 'go get')
	GOBIN := $(shell go env GOPATH | sed 's/:.*//')/bin
endif

tools = $(addprefix $(GOBIN)/, golangci-lint goimports govulncheck protoc-gen-go protoc-gen-go-grpc gci)
deps = $(addprefix $(GOBIN)/, wire dbmate)

ifneq (,$(wildcard ./.local.env))
    include ./.local.env
    export
endif

.DEFAULT_GOAL := all

###############################################################################
#
# Initialization
#
###############################################################################

init: ## Install git hooks to perform pre-commit checks
	git config core.hooksPath .githooks
	git config commit.template .gitmessage

dep: $(deps) ## Install the deps required to generate code and build Harness
	@echo "Installing dependencies"
	@go mod download

tools: $(tools) ## Install tools required for the build
	@echo "Installed tools"

###############################################################################
#
# Harness Build and testing rules
#
###############################################################################

build: generate ## Build the all-in-one Harness binary
	@echo "Building Harness Server"
	go build -o ./gitness ./cmd/gitness

test: generate  ## Run the go tests
	@echo "Running tests"
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out



###############################################################################
#
# Artifact Registry Build and testing rules
#
###############################################################################

run: clean build
	./gitness server .local.env || true

ar-conformance-test: clean build
	./gitness server .local.env > logfile.log 2>&1 & echo $$! > server.PID
	@sleep 10
	./registry/tests/conformance_test.sh localhost:3000 || true
	kill `cat server.PID`
	@rm server.PID
	@rm logfile.log

ar-hot-conformance-test:
	rm -rf distribution-spec || true
	./registry/tests/conformance_test.sh localhost:3000 || true

ar-api-update:
	@set -e; \
	oapi-codegen --config ./registry/config/openapi/artifact-services.yaml ./registry/app/api/openapi/api.yaml; \
	oapi-codegen --config ./registry/config/openapi/artifact-types.yaml ./registry/app/api/openapi/api.yaml;

ar-clean:
	@rm artifact-registry 2> /dev/null || true
	@docker stop ps_artifacthub 2> /dev/null || true
	rm -rf distribution-spec
	@kill -9 $$(lsof -t -i:3000) || true
	@rm server.PID || true
	@rm logfile.log || true
	go clean

###############################################################################
#
# Code Formatting and linting
#
###############################################################################

format: tools # Format go code and error if any changes are made
	@echo "Formatting ..."
	@goimports -w .
	@gci write --skip-generated --custom-order -s standard -s "prefix(github.com/harness/gitness)" -s default -s blank -s dot .
	@echo "Formatting complete"

sec:
	@echo "Vulnerability detection $(1)"
	@govulncheck ./...

lint: tools generate # lint the golang code
	@echo "Linting $(1)"
	@golangci-lint run --timeout=3m --verbose

###############################################################################
# Code Generation
#
# Some code generation can be slow, so we only run it if
# the source file has changed.
###############################################################################

generate: wire
	@echo "Generated Code"

wire: cmd/gitness/wire_gen.go

force-wire: ## Force wire code generation
	@sh ./scripts/wire/gitness.sh

cmd/gitness/wire_gen.go: cmd/gitness/wire.go
	@sh ./scripts/wire/gitness.sh

###############################################################################
# Install Tools and deps
#
# These targets specify the full path to where the tool is installed
# If the tool already exists it wont be re-installed.
###############################################################################

update-tools: delete-tools $(tools) ## Update the tools by deleting and re-installing

delete-tools: ## Delete the tools
	@rm $(tools) || true

# Install golangci-lint
$(GOBIN)/golangci-lint:
	@echo "🔘 Installing golangci-lint... (`date '+%H:%M:%S'`)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v1.56.2

# Install goimports to format code
$(GOBIN)/goimports:
	@echo "🔘 Installing goimports ... (`date '+%H:%M:%S'`)"
	@go install golang.org/x/tools/cmd/goimports

# Install wire to generate dependency injection
$(GOBIN)/wire:
	go install github.com/google/wire/cmd/wire@latest

# Install dbmate to perform db migrations
$(GOBIN)/dbmate:
	go install github.com/amacneil/dbmate@v1.15.0

$(GOBIN)/govulncheck:
	go install golang.org/x/vuln/cmd/govulncheck@v1.1.1

$(GOBIN)/protoc-gen-go:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28

$(GOBIN)/protoc-gen-go-grpc:
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

$(GOBIN)/gci:
	go install github.com/daixiang0/gci@v0.13.1

help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: delete-tools update-tools help format lint