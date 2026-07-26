# Task 035: Scaffold the Web UI as a Proper SPA

## Phase
3

## Type
feature

## Priority
P1

## Description
Replace the single hand-written `ui/dist/index.html` with a real single-page app
using a modern build toolchain, wired to the existing REST + WebSocket API served
by `labctl ui`. This task is the scaffold + API client only; views come in 036.

## Files to Modify
- `ui/` (new project: `package.json`, `vite.config.*`, `src/`, `tsconfig.json`)
- `.gitignore` (already ignores `ui/node_modules`, `ui/dist/*`)
- `docs/runbooks/03-web-ui.md`

## Implementation Notes
- Stack: **Vite + React + TypeScript** (or Svelte if preferred — pick one and be
  consistent). Keep dependencies lean; this is a dashboard, not an app store.
- Dev mode: Vite dev server with a proxy to `http://localhost:3939` for `/api` and
  the WebSocket endpoint, so `npm run dev` works against a running `labctl ui`.
- Build output must go to `ui/dist` (consumed by `go:embed` in task 037).
- Create a thin API client module wrapping the existing endpoints
  (`GET /api/status`, `/api/scenarios`, `POST /api/scenarios/:name/up`, etc. — see
  `cmd/labctl/internal/api/handlers.go` for the current surface) and a WebSocket
  hook for streamed `ActionEvent`s.
- Keep the dark theme aesthetic of the current dashboard.

## Acceptance Criteria
- [ ] `cd ui && npm install && npm run dev` serves the SPA and proxies to the API.
- [ ] `npm run build` emits a production bundle into `ui/dist`.
- [ ] API client + WebSocket hook exist and successfully fetch `/api/status`.
- [ ] No application code lives in a single inline HTML file anymore.

## Testing Instructions
Start `bin/labctl ui`, then `cd ui && npm run dev`; the dashboard loads cluster
status from the live API. Follow `docs/runbooks/03-web-ui.md`.

## Dependencies
None (but most useful after Phase 2 stabilizes the API)
