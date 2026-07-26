# Task 029: Validate PROFILE and APP_NAME Exist Before Use

## Priority
P1

## Assigned To
Bug

## Description
`cmd/labctl/internal/config/config.go` reads `PROFILE` from environment and constructs a path to `runtimes/${PROFILE}/runtime.env`, but never checks whether that directory exists. If `PROFILE` is set to a typo (e.g., `k3dd`, `eks2`), all `labctl` commands silently proceed with a partially initialized config — no runtime-specific env vars are loaded, `DOMAIN_SUFFIX` defaults to empty string, and subsequent operations produce confusing downstream errors (e.g., ingress hostnames become `app.` with no TLD). The same problem occurs with `APP_NAME` when building/deploying apps that don't exist in `apps/`.

## Files to Modify
- `cmd/labctl/internal/config/config.go`

## Implementation Notes
In the `Load()` function, after resolving `ProjectRoot` and `Profile`, add an existence check for the runtime directory:

```go
runtimeDir := filepath.Join(projectRoot, "runtimes", profile)
if _, err := os.Stat(runtimeDir); os.IsNotExist(err) {
    return nil, fmt.Errorf("runtime profile %q not found in runtimes/; available profiles: %s",
        profile, availableProfiles(projectRoot))
}
```

Implement `availableProfiles(root string) string` as a helper that reads the `runtimes/` directory and returns a comma-separated list of valid profile names.

In `LoadAppConfig()`, add a similar check for the app directory:

```go
appDir := filepath.Join(projectRoot, "apps", appName)
if _, err := os.Stat(appDir); os.IsNotExist(err) {
    return nil, fmt.Errorf("app %q not found in apps/; available apps: %s",
        appName, availableApps(projectRoot))
}
```

Return errors from both checks as typed errors so callers can distinguish "config not found" from "parse error". Both checks must list the valid options in the error message to make them immediately actionable.

## Acceptance Criteria
- [ ] `labctl --profile k3dd status` exits with a clear error: `runtime profile "k3dd" not found in runtimes/; available profiles: aks, eks, k3d`.
- [ ] `labctl app build nonexistent` exits with a clear error listing valid apps.
- [ ] Valid profiles and app names succeed as before with zero behavior change.
- [ ] `go test ./internal/config/...` passes with tests covering invalid profile and invalid app name.

## Testing Instructions
Run `PROFILE=bogus labctl status` and confirm the error message names the invalid profile and lists alternatives. Run `labctl app build fakeapp` and confirm the error message lists real apps.

## Dependencies
None
