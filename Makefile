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
BENCH_SERVICES ?= db-primary db-replica1 db-replica2 minio minio-init api api-read1 api-read2 node-exporter prometheus
HALSTEAD_MAX_VOLUME ?= 4000

ALLURE_RESULTS_DIR := allure-results
ALLURE_REPORT_DIR  := allure-report
ALLURE_HISTORY_DIR := allure-history
ALLURE_BRIDGE_CMD := go run ./cmd/allurebridge

TEST_PHASE ?= unit
E2E_REQUIRE_INFRA ?= 1

.PHONY: all clean dirs lint install-hooks \
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
	go test -tags=e2e -count=1 ./internal/tests/e2e

.PHONY: test-e2e-catfacts-mock
test-e2e-catfacts-mock:
	@echo "=> Running fun-fact e2e against standalone mock CatFacts server"
	@bash -c 'set -euo pipefail; \
	  addr=$${CATFACTS_ADDR:-:9099}; \
	  if [ -n "${CATFACTS_BASE_URL-}" ]; then \
	    base_url=$${CATFACTS_BASE_URL}; \
	  else \
	    case "$$addr" in \
	      http://*|https://*) base_url=$$addr ;; \
	      :*) base_url=http://localhost$$addr ;; \
	      *) base_url=http://$$addr ;; \
	    esac; \
	  fi; \
	  srv=""; \
	  trap "test -n \"$$srv\" && kill $$srv" EXIT; \
	  go run ./cmd/mockcatfacts -addr $$addr >/dev/null 2>&1 & srv=$$!; \
	  sleep 0.2; \
	  EXTERNAL_CATFACTS_MODE=mock EXTERNAL_CATFACTS_MOCK_BASE_URL=$$base_url \
	  go test -tags=e2e -count=1 -run TestE2E_DatasetFunFactFromExternalService ./internal/tests/e2e; \
	'

.PHONY: test-e2e-catfacts-real
test-e2e-catfacts-real:
	@echo "=> Running fun-fact e2e against real CatFacts service"
	EXTERNAL_CATFACTS_MODE=real E2E_REQUIRE_INFRA=$(E2E_REQUIRE_INFRA) \
		go test -tags=e2e -count=1 -run TestE2E_DatasetFunFactFromExternalService ./internal/tests/e2e

.PHONY: test-e2e-openai-mock
test-e2e-openai-mock:
	@echo "=> Running LLM summary e2e against standalone mock OpenAI server"
	@bash -c 'set -euo pipefail; \
	  addr=$${OPENAI_ADDR:-:8088}; \
	  case "$$addr" in \
	    http://*|https://*) base_url=$$addr ;; \
	    :*) base_url=http://localhost$$addr ;; \
	    *) base_url=http://$$addr ;; \
	  esac; \
	  srv=""; \
	  trap "test -n \"$$srv\" && kill $$srv" EXIT; \
	  go run ./cmd/mockopenai -addr $$addr >/dev/null 2>&1 & srv=$$!; \
	  sleep 0.2; \
	  EXTERNAL_OPENAI_MODE=mock EXTERNAL_OPENAI_BASE_URL=$$base_url \
	  go test -tags=e2e -count=1 -run TestE2E_DatasetSummaryFromOpenAI ./internal/tests/e2e; \
	'

.PHONY: test-e2e-openai-real
test-e2e-openai-real:
	@echo "=> Running LLM summary e2e against real OpenAI service"
	EXTERNAL_OPENAI_MODE=real E2E_REQUIRE_INFRA=$(E2E_REQUIRE_INFRA) \
		go test -tags=e2e -count=1 -run TestE2E_DatasetSummaryFromOpenAI ./internal/tests/e2e

.PHONY: test-e2e-2fa
test-e2e-2fa:
	@echo "=> Running 2FA BDD e2e scenario"
	E2E_REQUIRE_INFRA=$(E2E_REQUIRE_INFRA) go test -tags=e2e -count=1 -run TwoFactorFeatures ./internal/tests/e2e

.PHONY: lint
lint:
	golangci-lint run ./... --timeout=5m
	go run ./cmd/halsteadcheck -max-volume=4000

.PHONY: install-hooks
install-hooks:
	@echo "=> Installing git hooks"
	@ln -sf $(CURDIR)/scripts/pre-commit.sh .git/hooks/pre-commit

.PHONY: top_cycle
top_cycle:
	GOCACHE=$$(mktemp -d) go list -f '{{range .GoFiles}}{{$$.Dir}}/{{.}} {{end}}' ./internal/... | xargs gocyclo -top 10

.PHONY: top_halsted
top_halsted:
	@GOCACHE=$$(mktemp -d) go run ./cmd/halsteadcheck -max-volume=0 ./internal 2>&1 \
	| sed -n 's/^- //p' \
	| sort -t= -k2 -nr \
	| head -5

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

.PHONY: build-mockcatfacts
build-mockcatfacts: dirs
	@echo "=> Building CatFacts mock service (black-box artifact)"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/mockcatfacts ./cmd/mockcatfacts

.PHONY: build-mockopenai
build-mockopenai: dirs
	@echo "=> Building OpenAI mock service (black-box artifact)"
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/mockopenai ./cmd/mockopenai

.PHONY: info
info:
	@echo "Module: $(MODULE)"
	@echo "Build directory: $(BUILD)"
	@echo "CLI binary: $(BIN_DIR)/cli"
	@echo "API binary: $(BIN_DIR)/api"
	@echo "GUI binary: $(BIN_DIR)/gui"
	@echo "Mock CatFacts binary: $(BIN_DIR)/mockcatfacts"
	@echo "Mock OpenAI binary: $(BIN_DIR)/mockopenai"

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

.PHONY: bench-upload-ramp-find
bench-upload-ramp-find:
	@echo "=> Run k6 upload find breakpoint (ramping-vus, upload.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    UPLOAD_PHASES=find \
	    FIND_START_VUS=$${FIND_START_VUS:-200} FIND_END_VUS=$${FIND_END_VUS:-350} FIND_STEP_VUS=$${FIND_STEP_VUS:-20} \
	    FIND_STEP_DURATION=$${FIND_STEP_DURATION:-45s} FIND_HOLD_DURATION=$${FIND_HOLD_DURATION:-90s} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/upload.js

.PHONY: bench-upload-ramp-steady
bench-upload-ramp-steady:
	@echo "=> Run k6 upload steady (ramping-vus, upload.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    UPLOAD_PHASES=steady \
	    STEADY_VUS=$${STEADY_VUS:-220} STEADY_RAMP_DURATION=$${STEADY_RAMP_DURATION:-60s} STEADY_HOLD_DURATION=$${STEADY_HOLD_DURATION:-10m} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/upload.js

.PHONY: bench-upload-ramp-overload
bench-upload-ramp-overload:
	@echo "=> Run k6 upload overload/recovery (ramping-vus, upload.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    UPLOAD_PHASES=overload \
		STEADY_VUS=$${STEADY_VUS:-200} STEADY_RAMP_DURATION=$${OVERLOAD_RAMP_DURATION:-30s} STEADY_HOLD_DURATION=$${OVERLOAD_HOLD_DURATION:-60s} \
	    OVERLOAD_VUS=$${OVERLOAD_VUS:-500} OVERLOAD_RAMP_DURATION=$${OVERLOAD_RAMP_DURATION:-15s} OVERLOAD_HOLD_DURATION=$${OVERLOAD_HOLD_DURATION:-15s} \
	    RECOVERY_VUS=$${RECOVERY_VUS:-200} RECOVERY_RAMP_DURATION=$${RECOVERY_RAMP_DURATION:-0s} RECOVERY_HOLD_DURATION=$${RECOVERY_HOLD_DURATION:-30m} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/upload.js

.PHONY: bench-login-ramp-find
bench-login-ramp-find:
	@echo "=> Run k6 login find breakpoint (ramping-vus, login.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    LOGIN_PHASES=find \
	    FIND_START_VUS=$${FIND_START_VUS:-300} FIND_END_VUS=$${FIND_END_VUS:-600} FIND_STEP_VUS=$${FIND_STEP_VUS:-20} \
	    FIND_STEP_DURATION=$${FIND_STEP_DURATION:-60s} FIND_HOLD_DURATION=$${FIND_HOLD_DURATION:-90s} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/login.js

.PHONY: bench-login-ramp-steady
bench-login-ramp-steady:
	@echo "=> Run k6 login steady (ramping-vus, login.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    LOGIN_PHASES=steady \
	    STEADY_VUS=$${STEADY_VUS:-340} STEADY_RAMP_DURATION=$${STEADY_RAMP_DURATION:-60s} STEADY_HOLD_DURATION=$${STEADY_HOLD_DURATION:-10m} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/login.js

.PHONY: bench-login-ramp-overload
bench-login-ramp-overload:
	@echo "=> Run k6 login overload/recovery (ramping-vus, login.js)"
	docker compose up -d $(BENCH_SERVICES)
	env K6_PROMETHEUS_RW_SERVER_URL=http://localhost:9090/api/v1/write \
	    K6_PROMETHEUS_RW_TAGS_AS_LABELS=true \
	    K6_PROMETHEUS_RW_TREND_STATS="min,avg,med,p(75),p(90),p(95),p(99),max" \
	    LOGIN_PHASES=overload \
		STEADY_VUS=$${STEADY_VUS:-320} STEADY_RAMP_DURATION=$${STEADY_RAMP_DURATION:-30s} STEADY_HOLD_DURATION=$${STEADY_HOLD_DURATION:-90s} \
	    OVERLOAD_VUS=$${OVERLOAD_VUS:-600} OVERLOAD_RAMP_DURATION=$${OVERLOAD_RAMP_DURATION:-15s} OVERLOAD_HOLD_DURATION=$${OVERLOAD_HOLD_DURATION:-15s} \
	    RECOVERY_VUS=$${RECOVERY_VUS:-320} RECOVERY_RAMP_DURATION=$${RECOVERY_RAMP_DURATION:-0s} RECOVERY_HOLD_DURATION=$${RECOVERY_HOLD_DURATION:-30m} \
	    BASE_URL=$${BASE_URL:-http://localhost:8080} \
	    ./k6 run $${K6_RUN_ARGS:-} -o experimental-prometheus-rw scripts/k6/login.js
.PHONY: clean-storage
clean-storage:
	@echo "=> Clean MinIO bucket (mybucket)"
	./scripts/k6/clean_storage.sh
