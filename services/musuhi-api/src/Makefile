SHELL := /bin/bash

EXPORT_SCRIPT := ../../../tools/test_result_export/export_test_result.sh

.PHONY: help export-test-result template-test-result unit-result internal-result api-result e2e-result security-result performance-result regression-result external-template acceptance-template

help:
	@echo "Targets:"
	@echo "  make export-test-result TYPE=<type> TARGET_ID=<id> TEST_COMMAND='<command ...>'"
	@echo "  make template-test-result TYPE=<type> TARGET_ID=<id>"
	@echo "  make unit-result TARGET_ID=FR-001 TEST_FILTER='SystemOverview'"
	@echo "  make internal-result TARGET_ID=FR-001 TEST_FILTER='SystemOverview|ヘルスチェック'"
	@echo "  make api-result TARGET_ID=FR-001 TEST_COMMAND='go test ./test/integration -run API -v'"
	@echo "  make e2e-result TARGET_ID=SPR-1-1 TEST_COMMAND='npx playwright test'"
	@echo "  make security-result TARGET_ID=SPR-1-1 TEST_COMMAND='trivy image your-image:tag'"
	@echo "  make performance-result TARGET_ID=PH-1 TEST_COMMAND='k6 run script.js'"
	@echo "  make regression-result TARGET_ID=MRG-123 TEST_COMMAND='go test ./... -v'"
	@echo "  make external-template TARGET_ID=FR-001"
	@echo "  make acceptance-template TARGET_ID=FR-001"

export-test-result:
	@test -n "$(TYPE)" || (echo "TYPE is required" >&2; exit 1)
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) "$(TYPE)" "$(TARGET_ID)" $(TEST_COMMAND)

template-test-result:
	@test -n "$(TYPE)" || (echo "TYPE is required" >&2; exit 1)
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) --template "$(TYPE)" "$(TARGET_ID)"

unit-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_FILTER)" || (echo "TEST_FILTER is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) unit "$(TARGET_ID)" go test ./internal/handler ./internal/service -run "$(TEST_FILTER)" -v

internal-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_FILTER)" || (echo "TEST_FILTER is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) internal "$(TARGET_ID)" go test ./test/integration -run "$(TEST_FILTER)" -v

api-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) api "$(TARGET_ID)" $(TEST_COMMAND)

e2e-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) e2e "$(TARGET_ID)" $(TEST_COMMAND)

security-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) security "$(TARGET_ID)" $(TEST_COMMAND)

performance-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) performance "$(TARGET_ID)" $(TEST_COMMAND)

regression-result:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@test -n "$(TEST_COMMAND)" || (echo "TEST_COMMAND is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) regression "$(TARGET_ID)" $(TEST_COMMAND)

external-template:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) --template external "$(TARGET_ID)"

acceptance-template:
	@test -n "$(TARGET_ID)" || (echo "TARGET_ID is required" >&2; exit 1)
	@$(EXPORT_SCRIPT) --template acceptance "$(TARGET_ID)"
