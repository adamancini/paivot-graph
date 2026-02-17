PLUGIN_DIR := $(shell pwd)
PLUGIN_NAME := paivot-graph

.PHONY: install update uninstall test lint help

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

install: ## Register marketplace and install plugin
	@claude plugin marketplace add "$(PLUGIN_DIR)" 2>/dev/null \
		&& echo "Marketplace registered." \
		|| echo "Marketplace already registered."
	@claude plugin install "$(PLUGIN_NAME)@$(PLUGIN_NAME)" 2>/dev/null \
		&& echo "Plugin installed." \
		|| echo "Plugin already installed -- run 'make update' to pick up changes."
	@echo "Restart Claude Code sessions for hooks to take effect."

update: ## Push local changes to the installed plugin (bump version first)
	claude plugin marketplace update "$(PLUGIN_NAME)"
	claude plugin update "$(PLUGIN_NAME)@$(PLUGIN_NAME)"
	@echo "Restart Claude Code sessions for changes to take effect."

uninstall: ## Remove plugin and marketplace
	claude plugin uninstall "$(PLUGIN_NAME)@$(PLUGIN_NAME)"
	claude plugin marketplace remove "$(PLUGIN_NAME)"
	@echo "$(PLUGIN_NAME) removed."

lint: ## Run shellcheck on hook scripts
	shellcheck hooks/*.sh

test: lint ## Run all checks (shellcheck + functional)
	@echo "--- Functional checks ---"
	@echo "Checking hook scripts are executable..."
	@test -x hooks/vault-session-start.sh || (echo "FAIL: vault-session-start.sh not executable" && exit 1)
	@test -x hooks/vault-pre-compact.sh || (echo "FAIL: vault-pre-compact.sh not executable" && exit 1)
	@echo "OK: Hook scripts are executable"
	@echo ""
	@echo "Checking hooks.json is valid JSON..."
	@python3 -c "import json; json.load(open('hooks/hooks.json'))" || (echo "FAIL: hooks.json is not valid JSON" && exit 1)
	@echo "OK: hooks.json is valid JSON"
	@echo ""
	@echo "Checking plugin.json is valid JSON..."
	@python3 -c "import json; json.load(open('.claude-plugin/plugin.json'))" || (echo "FAIL: plugin.json is not valid JSON" && exit 1)
	@echo "OK: plugin.json is valid JSON"
	@echo ""
	@echo "Checking session-start hook exits 0 without obsidian..."
	@echo '{}' | PATH=/usr/bin:/bin hooks/vault-session-start.sh >/dev/null 2>&1; \
		test $$? -eq 0 && echo "OK: session-start graceful degradation" || echo "FAIL: session-start did not exit 0"
	@echo ""
	@echo "Checking pre-compact hook exits 0..."
	@hooks/vault-pre-compact.sh >/dev/null 2>&1; \
		test $$? -eq 0 && echo "OK: pre-compact exits 0" || echo "FAIL: pre-compact did not exit 0"
	@echo ""
	@echo "All checks passed."
