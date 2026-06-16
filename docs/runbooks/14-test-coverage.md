# Runbook 14: Running the Test Suite and Coverage Report

## Overview

Flightdeck uses `go test` for unit and integration tests. Tests are hermetic:
no live cluster or cloud credentials required. Coverage must stay at or above
75% for any package added or significantly modified (golden rule 10).

## Running the test suite

```bash
# From the repo root
cd cmd/labctl && go test ./...
```

All packages should report `ok`. No `FAIL` lines means the suite is green.

## Running with coverage

```bash
# Per-package coverage percentages (quick check)
cd cmd/labctl && go test -cover ./...

# HTML report (opens in a browser after the command)
make test-coverage
open coverage.html      # macOS
xdg-open coverage.html  # Linux
```

`make test-coverage` writes `coverage.out` (raw profile) and `coverage.html`
(annotated source) to the repo root.

## Interpreting the HTML report

- Green lines are covered; red lines are not.
- Click a file in the top-left dropdown to navigate.
- Focus on uncovered branches in error-handling paths — those are the most
  valuable to add.

## Coverage thresholds

| Target | Threshold |
|--------|-----------|
| Any new or significantly modified package | ≥ 75% |
| Engine packages (scenario, incident, challenge) | ≥ 75% |
| API handlers | best-effort (live cluster paths are excluded) |

## Adding tests

Follow the patterns in existing `*_test.go` files:

- Use `t.TempDir()` for all disk I/O — never write to a fixed path.
- Use table-driven tests (`tests := []struct{...}{}`) for 2+ similar cases.
- Mock external calls; never depend on a real cluster, network, or real scripts.
- Place test helpers in the same `package foo` (not `package foo_test`) so
  private functions are accessible.

## Checking a specific package

```bash
cd cmd/labctl
go test -coverprofile=/tmp/cov.out ./internal/scenario/...
go tool cover -func=/tmp/cov.out | grep -v "100.0%"
```

This shows exactly which functions are below full coverage.
