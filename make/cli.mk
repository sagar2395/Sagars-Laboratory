CLI_DIR     := cmd/labctl
CLI_BIN     := bin/labctl
CLI_MODULE  := github.com/sagars-lab/labctl
CLI_UI_SRC  := ui/dist
CLI_UI_DEST := $(CLI_DIR)/ui/dist

.PHONY: cli-build cli-install cli-tidy cli-clean

cli-build:
	@echo "Copying UI assets..."
	@cp -r $(CLI_UI_SRC)/* $(CLI_UI_DEST)/ 2>/dev/null || true
	@echo "Building labctl..."
	@cd $(CLI_DIR) && go build -o ../../$(CLI_BIN) .
	@echo "Binary: $(CLI_BIN)"

cli-tidy:
	@cd $(CLI_DIR) && go mod tidy

cli-install: cli-build
	@cp $(CLI_BIN) $(GOPATH)/bin/labctl 2>/dev/null || cp $(CLI_BIN) /usr/local/bin/labctl
	@echo "Installed labctl to PATH"

cli-clean:
	@rm -f $(CLI_BIN)
	@rm -f $(CLI_UI_DEST)/index.html
