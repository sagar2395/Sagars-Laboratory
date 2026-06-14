# Task 079: Comprehensive Test Suite Audit & Gap Fill

## Phase
Part II/III — cross-cutting quality gate (runs before any M5+ feature work)

## Type
review / infra

## Priority
P0

## Description

The project ships 46 test files (~7,730 lines) written as the features were built.
Before accelerating M5/M6 work, audit every package for coverage gaps, fill the
most critical ones, and add a `make test-coverage` target so coverage is visible in CI.
This task does NOT add new features — it only adds tests for things that already exist.

**Why now:** AI sessions generating M5+ code will extend packages that already have
tests. If those existing tests have blind spots in error paths, edge cases, or
integration flows, regressions will be invisible until they reach production.

### Scope of audit (priority order)

| Priority | Package(s) | Known gap |
|----------|------------|-----------|
| P0 | `pkg/checks` | Happy path covered; timeout/cancellation/parallel-failure edge cases incomplete |
| P0 | `internal/scenario` (engine + catalog) | Schema validation errors and v2→v1 fallback not tested |
| P0 | `internal/incident` | Hint-penalty accumulation + MTTR calculation boundary values |
| P0 | `internal/challenge` | gradeDetection when check script exits non-zero mid-run |
| P1 | `internal/executor` | Cancellation propagation; script writing racy teardown |
| P1 | `internal/learn` | Completion when last lesson has no checks |
| P1 | `internal/platform` | registry.Resolve with unknown provider (error path) |
| P1 | `pkg/pack` (oci + index) | Unsigned image verify path; index search with no results |
| P2 | `cmd/labctl/cmd/` | CLI command smoke-tests (no unit tests exist for cobra commands) |
| P2 | `internal/api/` | Handler error responses (4xx/5xx) not asserted in most handlers |

---

## Files to Modify

### New test files
- `cmd/labctl/internal/scenario/engine_error_test.go` — error/edge cases for engine
- `cmd/labctl/internal/incident/mttr_edge_test.go` — MTTR + hint-penalty boundaries
- `cmd/labctl/internal/challenge/grade_error_test.go` — grading failure modes
- `cmd/labctl/internal/executor/cancel_test.go` — cancellation propagation
- `cmd/labctl/internal/learn/completion_edge_test.go` — edge cases for lesson completion
- `cmd/labctl/internal/platform/resolve_error_test.go` — unknown-provider errors
- `cmd/labctl/pkg/pack/index_edge_test.go` — empty/malformed index edge cases
- `cmd/labctl/internal/api/error_response_test.go` — 4xx/5xx API handler assertions
- `cmd/labctl/cmd/scenario_cmd_test.go` — cobra command smoke-tests

### Modified files
- `make/cli.mk` — add `test-coverage` target
- `Makefile` — wire `test-coverage` into the default help section

---

## Implementation Notes

### Test quality standards for this task
Every new test file in this task must:
1. Use **table-driven tests** (`tests := []struct{...}`) wherever two or more
   cases share the same shape — never repeat `TestFoo_A`, `TestFoo_B` as
   separate top-level functions for the same behaviour.
2. Test **at least one error/failure path** per public function touched.
3. Use `t.Helper()` on all test-helper functions.
4. Be **hermetic**: create a `t.TempDir()` for any disk I/O; never depend on
   the host's cluster, network, or external binaries.
5. Not import test-only helper packages that aren't already in `go.mod`.

### Coverage target
- Run `go test -coverprofile=coverage.out ./...` from `cmd/labctl/`.
- The goal for this task is **≥ 75% statement coverage** across all packages
  in `cmd/labctl/` (measured, not guessed).
- For any package whose coverage is below 50% after this task, add a
  `// TODO(test): coverage XX% — needs table tests for <description>` comment
  at the top of the existing `_test.go` file as a signal for future sessions.

### make test-coverage target
```makefile
.PHONY: test-coverage
test-coverage: ## Run tests with HTML coverage report (opens in browser on macOS)
	cd cmd/labctl && go test -coverprofile=../../coverage.out ./... && \
	  go tool cover -html=../../coverage.out -o ../../coverage.html
	@echo "Coverage report: coverage.html"
```

### Cobra command tests
Use `cobra`'s `ExecuteC()` with a test root command. Do NOT spin up a cluster;
stub out the labctl config file and assert that `--help` exits 0 and key flags
are registered. Example pattern:
```go
func TestScenarioCmdFlags(t *testing.T) {
    cmd := newRootCmd()
    cmd.SetArgs([]string{"scenario", "--help"})
    err := cmd.Execute()
    if err != nil { t.Fatal(err) }
}
```

---

## Acceptance Criteria

- [ ] `cd cmd/labctl && go test ./...` passes with no new failures.
- [ ] `make test-coverage` target exists and produces `coverage.html`.
- [ ] Statement coverage for `cmd/labctl/` is ≥ 75% (printed by the make target).
- [ ] Each of the P0 packages listed above has at least 3 new table-driven cases
      covering error/edge paths that were not previously tested.
- [ ] The P1 packages each have at least 1 new negative test (error path).
- [ ] At least 5 cobra CLI command smoke-tests exist in `cmd/labctl/cmd/`.
- [ ] `docs/runbooks/14-test-coverage.md` exists and shows how to run the suite
      and interpret the HTML report.

---

## Testing Instructions

```bash
# Run the full test suite
cd cmd/labctl && go test ./... -v 2>&1 | tail -40

# Run with coverage
make test-coverage

# Check a specific package
cd cmd/labctl && go test -v -cover ./internal/scenario/...
cd cmd/labctl && go test -v -cover ./internal/incident/...
cd cmd/labctl && go test -v -cover ./pkg/checks/...
```

Manual verification: open `coverage.html` in a browser; no red (uncovered) lines
should appear in the public-facing `pkg/` packages.

---

## Dependencies
None — purely additive test files on top of existing code.
