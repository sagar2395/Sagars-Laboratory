CLI_DIR     := cmd/labctl
CLI_BIN     := bin/labctl
CLI_MODULE  := github.com/sagars-lab/labctl
CLI_UI_SRC  := ui/dist
CLI_UI_DEST := $(CLI_DIR)/ui/dist

.PHONY: cli-build cli-build-all cli-install cli-tidy cli-clean

cli-build:
	@echo "Building UI..."
	@cd ui && npm install && npm run build
	@echo "Copying UI assets..."
<<<<<<< HEAD
<<<<<<< HEAD
	@mkdir -p $(CLI_UI_DEST)
	@cp -r ui/dist/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@echo "Building labctl..."
=======
	@cp -r $(CLI_UI_SRC)/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@echo "Building labctl for host ($(shell go env GOOS)/$(shell go env GOARCH))..."
>>>>>>> dd2d4c6 (Updating Phase 0, 1, 2 implementation)
=======
	@mkdir -p $(CLI_UI_DEST)
	@cp -r ui/dist/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@echo "Building labctl..."
>>>>>>> 9b97903 (Fixing conflict)
	@cd $(CLI_DIR) && go build -o ../../$(CLI_BIN) .
	@echo "Binary: $(CLI_BIN)"

# Cross-compile for all release targets. Outputs land in dist/.
cli-build-all:
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
