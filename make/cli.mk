CLI_DIR     := cmd/labctl
CLI_BIN     := bin/labctl
CLI_MODULE  := github.com/sagars-lab/labctl
CLI_UI_SRC  := ui/dist
CLI_UI_DEST := $(CLI_DIR)/ui/dist

.PHONY: cli-build cli-build-all cli-install cli-tidy cli-clean

cli-build:
	@echo "Building UI (SPA)..."
	@if command -v npm >/dev/null 2>&1; then \
		cd ui && npm ci --prefer-offline && npm run build && cd ..; \
	else \
		echo "Warning: npm not found — using pre-built UI assets in $(CLI_UI_SRC). Run 'npm ci && npm run build' in ui/ to refresh."; \
	fi
	@echo "Copying UI assets into embed target..."
	@cp -r $(CLI_UI_SRC)/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@echo "Building labctl for host ($(shell go env GOOS)/$(shell go env GOARCH))..."
	@cd $(CLI_DIR) && go build -o ../../$(CLI_BIN) .
	@echo "Binary: $(CLI_BIN)"

# Cross-compile for all release targets. Outputs land in dist/.
cli-build-all:
	@echo "Building UI (SPA)..."
	@if command -v npm >/dev/null 2>&1; then \
		cd ui && npm ci --prefer-offline && npm run build && cd ..; \
	else \
		echo "Warning: npm not found — using pre-built UI assets."; \
	fi
	@echo "Copying UI assets..."
	@cp -r $(CLI_UI_SRC)/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@mkdir -p dist
	@for target in darwin/arm64 darwin/amd64 linux/amd64 linux/arm64; do \
		goos=$${target%/*}; goarch=$${target#*/}; \
		out="dist/labctl-$${goos}-$${goarch}"; \
		echo "  Building $${out}..."; \
		cd $(CLI_DIR) && GOOS=$$goos GOARCH=$$goarch go build -o ../../$$out . && cd ../..; \
	done
	@echo "Cross-compiled binaries in dist/"

cli-tidy:
	@cd $(CLI_DIR) && go mod tidy

cli-install: cli-build
	@cp $(CLI_BIN) $(GOPATH)/bin/labctl 2>/dev/null || cp $(CLI_BIN) /usr/local/bin/labctl
	@echo "Installed labctl to PATH"

cli-clean:
	@rm -f $(CLI_BIN)
	@rm -rf dist/
	@rm -f $(CLI_UI_DEST)/index.html

.PHONY: test-coverage
test-coverage: ## Run tests with HTML coverage report
	cd $(CLI_DIR) && go test -coverprofile=../../coverage.out ./... && \
	  go tool cover -html=../../coverage.out -o ../../coverage.html
	@echo "Coverage report: coverage.html"
