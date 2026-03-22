# Task 012: Fix Engine Strategy Script Existence Validation

## Priority
P1

## Assigned To
Bug

## Description
`engine/build.sh` and `engine/deploy.sh` source a strategy script path derived from `BUILD_STRATEGY` and `DEPLOY_STRATEGY` variables in `app.env`. If the resolved path does not exist (e.g., a typo in the env file, or a strategy removed from the repo), Bash sources a non-existent file and produces a confusing error message deep in the stack rather than a clear early failure. This must be caught at the dispatch layer with a meaningful message before any work begins.

## Files to Modify
- `engine/build.sh`
- `engine/deploy.sh`

## Implementation Notes
After resolving the strategy script path and before sourcing/executing it, add an existence and executability check:

```bash
if [[ ! -f "${STRATEGY_SCRIPT}" ]]; then
  echo "[engine] ERROR: Strategy script not found: ${STRATEGY_SCRIPT}" >&2
  echo "[engine] Check BUILD_STRATEGY (or DEPLOY_STRATEGY) in app.env" >&2
  exit 1
fi

if [[ ! -x "${STRATEGY_SCRIPT}" ]]; then
  echo "[engine] ERROR: Strategy script is not executable: ${STRATEGY_SCRIPT}" >&2
  exit 1
fi
```

Do not change any execution logic, timeouts, or output formatting beyond these guard clauses.

## Acceptance Criteria
- [ ] Setting `BUILD_STRATEGY` to a non-existent path causes `engine/build.sh` to exit 1 with a clear error message identifying the missing file.
- [ ] Setting `DEPLOY_STRATEGY` to a non-existent path causes `engine/deploy.sh` to exit 1 similarly.
- [ ] A non-executable strategy script also produces a clear error (not a permission denied from Bash internals).
- [ ] No behavior changes when strategy scripts exist and are executable.

## Testing Instructions
In a test `app.env`, set `BUILD_STRATEGY=engine/build/nonexistent.sh` and run `make app-build APP=go-api`. Confirm the error message includes the full path and a hint about `app.env`. Repeat for deploy. Then restore the correct strategy and confirm normal operation.

## Dependencies
None
