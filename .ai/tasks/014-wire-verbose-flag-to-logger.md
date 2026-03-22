# Task 014: Wire --verbose Flag to Structured Logger

## Priority
P2

## Assigned To
Feature

## Description
The `labctl` root command defines a `--verbose` / `-v` persistent flag and stores its value in a config struct, but the flag is never read by the logger initialization. All log output is currently at the same verbosity level regardless of whether `--verbose` is passed. This makes debugging failures significantly harder since there is no way to get more diagnostic output without modifying source code.

## Files to Modify
- `cmd/labctl/cmd/root.go`
- `cmd/labctl/cmd/ui.go`

## Implementation Notes
1. In `root.go`, after parsing the `--verbose` flag, check its value and configure the `slog` level accordingly:
   - Default: `slog.LevelWarn` (or `slog.LevelInfo` if that is the current default — preserve existing behavior).
   - `--verbose`: `slog.LevelDebug`.
2. Use `slog.SetLogLoggerLevel` or replace the default handler with a `slog.NewTextHandler` at the appropriate level.
3. In relevant subcommands (executor calls, script runs, API client calls), add `slog.Debug(...)` calls at key boundaries: before/after script execution, on HTTP request/response, on config load. You do not need to add debug logging everywhere — focus on the 3-5 most useful diagnostic points.
4. Update the root command's help text to mention what `--verbose` enables.

Do not add new dependencies. Use the standard library `log/slog` only.

## Acceptance Criteria
- [ ] `labctl --verbose <subcommand>` produces debug-level output including key execution boundaries.
- [ ] `labctl <subcommand>` (without `--verbose`) produces the same output as before this change.
- [ ] `--verbose` appears with a useful description in `labctl --help`.
- [ ] At minimum, script execution start/end and API call details are logged at debug level.

## Testing Instructions
Run `labctl --verbose platform status` and confirm additional diagnostic lines appear compared to `labctl platform status`. Run `go test ./cmd/...` to confirm no regressions.

## Dependencies
None
