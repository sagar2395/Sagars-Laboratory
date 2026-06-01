# Runbook 03 — Web UI

> Goal: build, run, and drive the dashboard. Verifies Phase 3 (SPA rebuild).

## Prereqs

- Runbook 00 (cluster up). Node.js 20+ for frontend dev.

## Frontend dev loop

```bash
# Terminal 1 — API + (embedded) UI server
bin/labctl ui                       # serves on http://localhost:3939

# Terminal 2 — live-reloading SPA dev server (proxies /api + WS to :3939)
cd ui
npm install
npm run dev                         # open the printed localhost URL
```

## Production build (embedded into the binary)

```bash
make cli-build                      # builds SPA -> ui/dist -> embeds via go:embed
bin/labctl ui                       # serves the embedded SPA at :3939
```

## Exercise each view

1. **Dashboard** — confirm cluster info (runtime, nodes, k8s version, context) and
   the platform health grid match `bin/labctl status` / `platform status`.
2. **Apps** — deploy `go-api`, watch pod status update, open its logs link.
3. **Scenarios** — start `observability-sre`; progress streams live; it then shows
   as active. Stop it; it returns to inactive.
4. **Platform** — switch the ingress provider; the action streams and completes.

## Expected

- `bin/labctl ui` serves the SPA with **no external files** (move/rename `ui/dist`
  and it still loads — proves embedding).
- Long actions stream live over WebSocket (depends on Task 003).
- Deep links (e.g. `/scenarios`) load directly (SPA fallback to `index.html`).

## Cleanup

Stop the `labctl ui` process (Ctrl-C). No cluster changes are made by viewing.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| Blank page on `bin/labctl ui` | SPA not built/embedded — `make cli-build` (Task 037). |
| 404 on deep link refresh | SPA fallback not configured in `server.go` (Task 037). |
| Live progress not updating | WebSocket not streaming all ops (Task 003). |
| `npm run dev` can't reach API | Start `bin/labctl ui` first; check Vite proxy target `:3939`. |
