MODULE := $(shell go list -m)

BUILD   := build
OBJ_DIR := $(BUILD)/obj
BIN_DIR := $(BUILD)/bin

DATA_PKG := ./internal/dataaccess/repositories/postqbuild
SVCS_PKG := ./internal/services
CLI_PKG  := ./cmd/cli
API_PKG  := ./cmd/api
GUI_PKG  := ./cmd/gui

GO     := go
GOTOOL := go tool
MKDIR  := mkdir -p
RM     := rm -rf


.PHONY: all clean dirs \
        build-dataaccess-archive build-services-archive \
        build-cli build-api build-gui info

all: dirs build-dataaccess-archive build-services-archive \
      build-cli build-api build-gui


.PHONY: dirs
dirs:
	$(MKDIR) $(OBJ_DIR)/dataaccess
	$(MKDIR) $(OBJ_DIR)/services
	$(MKDIR) $(BIN_DIR)

.PHONY: build-dataaccess-archive
build-dataaccess-archive: dirs
	@echo "=> Building DataAccess (postqbuild) as archive…"
	GO111MODULE=on $(GO) build -buildmode=archive \
	  -o $(OBJ_DIR)/dataaccess/dataaccess.a \
	  $(DATA_PKG)


.PHONY: build-services-archive
build-services-archive: dirs
	@echo "=> Building Services (internal/services) as archive…"
	GO111MODULE=on $(GO) build -buildmode=archive \
	  -o $(OBJ_DIR)/services/services.a \
	  $(SVCS_PKG)


.PHONY: build-cli
build-cli: build-dataaccess-archive build-services-archive
	@echo "=> Building CLI binary…"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/cli $(CLI_PKG)


.PHONY: build-api
build-api: build-dataaccess-archive build-services-archive
	@echo "=> Building API (Swagger) binary…"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/api $(API_PKG)

.PHONY: build-gui
build-gui: build-dataaccess-archive build-services-archive
	@echo "=> Building GUI binary…"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/gui $(GUI_PKG)

.PHONY: info
info:
	@echo "Module: $(MODULE)"
	@echo "Build directory: $(BUILD)"
	@echo "CLI binary: $(BIN_DIR)/cli"
	@echo "API binary: $(BIN_DIR)/api"
	@echo "GUI binary: $(BIN_DIR)/gui"

.PHONY: clean
clean:
	@echo "=> Cleaning build directory…"
	$(RM) $(BUILD)

.PHONY: clean-logs
clean-logs:
	@echo "=> Cleaning logs directory…"
	$(RM) logs/*