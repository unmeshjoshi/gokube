# Go parameters
GOTESTSUM=gotestsum
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOINSTALL=$(GOCMD) install

ifeq ($(shell command -v $(GOTESTSUM) > /dev/null 2>&1; echo $$?), 0)
	GOTEST=$(GOTESTSUM) --format testdox
else
	GOTEST=$(GOCMD) test -v
endif

# Make parameters
OUT_DIR=out
BINARIES=apiserver controller kubelet scheduler
BINARY_PATHS=$(addprefix $(OUT_DIR)/,$(BINARIES))
EXECUTABLES=$(addprefix $(GOPATH)/,$(BINARIES))

BUILD_TARGETS=$(addprefix build/,$(BINARIES))
INSTALL_TARGETS=$(addprefix install/,$(BINARIES))
GO_BIN_TARGETS=$(addprefix $(GOPATH)/bin/,$(BINARIES))

# Colors
CYAN_COLOR_START := \033[36m
CYAN_COLOR_END := \033[0m

.PHONY: precommit mockgen install clean test

help: ## Prints help (only for targets with comments)
	@grep -E '^[a-zA-Z._/\-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk -F'[:##]' '{printf "$(CYAN_COLOR_START)%-20s $(CYAN_COLOR_END)%s\n", $$2, $$5}'

deps: ## Install/Upgrade dependencies
	$(GOGET) ./...
	$(GOMOD) tidy

fmt: ## Format go code
	gofmt -s -w .

fmt-check: ## Check whether gofmt has been applied
	@if [ ! -z "$$(gofmt -l .)" ]; then \
		echo "Your code is not formatted. Run 'make fmt' to format the code"; \
		exit 1; \
	fi

vet: ## Run SCA using go vet
	go vet $(shell go list ./...)

lint: ## Run lint
	docker run --rm -v $(PWD):/app -v $(PWD)/.golangci-lint-cache:/root/.cache -w /app golangci/golangci-lint:v1.63.4 golangci-lint run -v --exclude S1000

test: ## Run all tests
	$(GOTEST) ./...

test/%: ## Run package level tests
	$(GOTEST) ./pkg/$(@F)

mockgen: install-mockgen ## Generate mocks using mockgen
	PROJECT_HOME=$(PWD) go generate ./...

install-mockgen: ## Install mockgen
	@if ! [ -x "$$(command -v mockgen)" ]; then \
		echo "mockgen not found, installing..."; \
		$(GOCMD) install go.uber.org/mock/mockgen@latest; \
	fi

$(OUT_DIR): ## Ensure output directory exists
	@if [ ! -d $(OUT_DIR) ]; then mkdir -p $(OUT_DIR); fi

$(OUT_DIR)/%: ## Build to out directory
	@$(GOBUILD) -o $(@) -v ./cmd/$(@F)/$(@F).go
	@printf "Built %s\n" $(@F)

build/apiserver: $(OUT_DIR)/apiserver ## Build apiserver
build/controller: $(OUT_DIR)/controller ## Build controller
build/kubelet: $(OUT_DIR)/kubelet ## Build kubelet
build/scheduler: $(OUT_DIR)/scheduler ## Build scheduler

build: build/apiserver build/controller build/kubelet build/scheduler ## Build all

precommit: deps fmt vet lint test build ## Run precommit target(deps,fmt,vet,lint,test)
	@echo "CI build completed successfully"

$(GO_BIN_TARGETS):
	@printf "Installing %s...\n" $(@F)
	@$(GOINSTALL) ./cmd/$(@F)/$(@F).go
	@printf "Successfully installed %s\n" $(@F)
	@printf "Executable located at %s\n\n" $(GOPATH)/bin/$(@F)

install/apiserver: $(GOPATH)/bin/apiserver ## Install apiserver in $(GOPATH)/bin
install/controller: $(GOPATH)/bin/controller ## Install controller in $(GOPATH)/bin
install/kubelet: $(GOPATH)/bin/kubelet ## Install kubelet in $(GOPATH)/bin
install/scheduler: $(GOPATH)/bin/scheduler ## Install scheduler in $(GOPATH)/bin

install: install/apiserver install/controller install/kubelet install/scheduler ## Install all
run: ### Run the project
	process-compose -f process-compose.yml up

clean: ## Cleans all directories
	@$(GOCLEAN)
	@rm -f $(BINARY_PATHS)
	@rm -rf $(OUT_DIR)
	@printf "Cleaned up build artifacts\n"
	@rm -f $(EXECUTABLES)
	@printf "Cleaned up installed binaries\n"
	@rm -rf $(DIST_DIR)
	@printf "Cleaned up dist artifacts\n"
	@rm -rf $(HOME)/gokube
	@printf "Cleaned up gokube binaries\n"

include limactl.mk
include colima.mk