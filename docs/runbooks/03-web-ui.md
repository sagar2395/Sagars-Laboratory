# Runbook 03: Web UI Validation & Backend Integration Notes

## Purpose
This runbook describes how to manually verify the Phase 3 SPA functionality, and serves as a contract guide for other AI agents integrating future backend features with the UI.

## Integration Notes for Backend AI Agents
If you are modifying the `labctl` Go backend, adhere to the following UI contracts:

1. **WebSocket Event Stream (`/api/events`)**
   - The UI subscribes to `ActionEvent` JSON messages to stream terminal logs and status updates for long-running operations (Scenarios, Platform components).
   - Expected shape: `{ "type": "action", "data": { "ID": string, "Type": string, "Action": string, "ExitCode": number, "Error": string, "Timestamp": string, "Line": string } }`
   - The UI expects a periodic `{ "type": "status", "data": ClusterInfo }` payload to keep the cluster status alive without HTTP polling.

2. **Routing & Embedding**
   - The SPA relies on `react-router-dom` for client-side routing. All non-API routes requested from the Go backend MUST fallback to serving `index.html` to avoid `404 Not Found` errors when users refresh deep links (e.g. `/scenarios`).
   - The Vite build output is located in `ui/dist` and embedded via `go:embed` into `cmd/labctl/ui/dist`. Do not change these paths without updating `make/cli.mk`.

3. **Status Polling**
   - For high-level status (Platform, Apps, Scenarios), the UI polls REST endpoints (`/api/status`, `/api/scenarios`, `/api/platform/status`, `/api/apps`) every 5 seconds. Avoid expensive operations in these handlers (use caching).

4. **Action Handlers**
   - Operations like `POST /api/scenarios/:name/up` are expected to return a `202 Accepted` immediately and kick off a background goroutine, emitting progress to the WebSocket stream.

## Manual Verification Steps
1. Rebuild the CLI with embedded UI: `make cli-build`.
2. Start the `labctl ui` server.
3. Open a browser to `http://localhost:3939`.
4. Verify the Dashboard loads with active cluster state.
5. Navigate to **Scenarios**, click "Start" on a scenario, and observe the built-in terminal stream WebSocket events in real-time.
6. Refresh the page on the `/scenarios` route to ensure the SPA fallback correctly serves the React application.
