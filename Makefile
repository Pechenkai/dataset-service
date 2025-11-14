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

ALLURE_RESULTS_DIR := allure-results
ALLURE_REPORT_DIR  := allure-report
ALLURE_HISTORY_DIR := allure-history
ALLURE_BRIDGE_CMD := go run ./cmd/allurebridge

TEST_PHASE ?= unit

.PHONY: all clean dirs \
        build-dataaccess-archive build-services-archive \
        build-cli build-api build-gui info \
        test test-unit test-integration test-e2e test-allure

all: dirs build-dataaccess-archive build-services-archive \
      build-cli build-api build-gui

.PHONY: test
test:
	@case "$(TEST_PHASE)" in \
	  unit) $(MAKE) test-unit ;; \
	  integration) $(MAKE) test-integration ;; \
	  e2e) $(MAKE) test-e2e ;; \
	  allure) $(MAKE) test-allure ;; \
	  *) echo "Unknown TEST_PHASE: $(TEST_PHASE)" >&2; exit 1 ;; \
	esac

.PHONY: test-unit
test-unit:
	@echo "=> Running unit tests"
	go test ./...

.PHONY: test-integration
test-integration:
	@echo "=> Running integration tests"
	go test -tags=integration ./internal/tests/access_tests/... ./internal/tests/integration_tests

.PHONY: test-e2e
test-e2e:
	@echo "=> Running e2e tests"
	go test -tags=e2e ./internal/tests/e2e

.PHONY: test-allure
test-allure: test-dataset-allure

.PHONY: dirs
dirs:
	$(MKDIR) $(OBJ_DIR)/dataaccess
	$(MKDIR) $(OBJ_DIR)/services
	$(MKDIR) $(BIN_DIR)

.PHONY: build-dataaccess-archive
build-dataaccess-archive: dirs
	@echo "=> Building DataAccess"
	GO111MODULE=on $(GO) build -buildmode=archive \
	  -o $(OBJ_DIR)/dataaccess/dataaccess.a \
	  $(DATA_PKG)


.PHONY: build-services-archive
build-services-archive: dirs
	@echo "=> Building Services"
	GO111MODULE=on $(GO) build -buildmode=archive \
	  -o $(OBJ_DIR)/services/services.a \
	  $(SVCS_PKG)


.PHONY: build-cli
build-cli: build-dataaccess-archive build-services-archive
	@echo "=> Building CLI"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/cli $(CLI_PKG)


.PHONY: build-api
build-api: build-dataaccess-archive build-services-archive
	@echo "=> Building API (Swagger)"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/api $(API_PKG)

.PHONY: build-gui
build-gui: build-dataaccess-archive build-services-archive
	@echo "=> Building GUI"
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

.PHONY: test-dataset-allure
test-dataset-allure:
	@echo "=> Running service Allure suites"
	@bash -c 'if [ -d $(ALLURE_RESULTS_DIR) ]; then find $(ALLURE_RESULTS_DIR) -mindepth 1 -delete; else mkdir -p $(ALLURE_RESULTS_DIR); fi'
	@bash -c 'if [ -d internal/tests/service_tests/$(ALLURE_RESULTS_DIR) ]; then find internal/tests/service_tests/$(ALLURE_RESULTS_DIR) -mindepth 1 -delete; else mkdir -p internal/tests/service_tests/$(ALLURE_RESULTS_DIR); fi'
	if [ -d $(ALLURE_HISTORY_DIR)/history ]; then \
	  cp -R $(ALLURE_HISTORY_DIR)/history $(ALLURE_RESULTS_DIR)/; \
	fi
	- GO_TEST_RUNNER=allure ALLURE_OUTPUT_PATH=$(CURDIR) $(GO) test -shuffle=on -p 1 ./internal/tests/service_tests -run Test.*ServiceSuite
	@echo "=> Capturing integration tests to Allure"
	@bash -c 'set -o pipefail; GOCACHE=$$(mktemp -d) $(GO) test -json -tags=integration ./internal/tests/access_tests/... ./internal/tests/integration_tests | $(ALLURE_BRIDGE_CMD) --suite integration --output $(CURDIR)/$(ALLURE_RESULTS_DIR)'
	@echo "=> Capturing e2e tests to Allure"
	@bash -c 'set -o pipefail; GOCACHE=$$(mktemp -d) $(GO) test -json -tags=e2e ./internal/tests/e2e | $(ALLURE_BRIDGE_CMD) --suite e2e --output $(CURDIR)/$(ALLURE_RESULTS_DIR)'

.PHONY: allure-report
allure-report: test-dataset-allure
	@echo "=> Generating Allure report"
	allure generate $(ALLURE_RESULTS_DIR) -o $(ALLURE_REPORT_DIR) --clean
	$(RM) -r $(ALLURE_HISTORY_DIR)
	$(MKDIR) $(ALLURE_HISTORY_DIR)
	if [ -d $(ALLURE_REPORT_DIR)/history ]; then \
	  cp -R $(ALLURE_REPORT_DIR)/history $(ALLURE_HISTORY_DIR)/; \
	fi

.PHONY: clean-logs
clean-logs:
	@echo "=> Cleaning logs directory…"
	$(RM) logs/*

.PHONY: e2e-capture
e2e-capture:
	@echo "=> Running HTTP scenario and capturing traffic"
	LOG_FILE=logs/e2e_capture_example.txt ./scripts/e2e_capture.sh

.PHONY: e2e-wireshark
e2e-wireshark:
	@echo "=> Running HTTP scenario with PCAP capture for Wireshark"
	LOG_FILE=logs/e2e_capture_example.txt PCAP_FILE=logs/e2e_capture.pcap ./scripts/e2e_capture.sh
	@echo "=> Open logs/e2e_capture.pcap in Wireshark to inspect the traffic"
