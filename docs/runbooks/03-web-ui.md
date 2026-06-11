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
   the platform health grid match `bin/labctl status` / `platform status`. The
   cluster card updates live (every 5 s over WebSocket) — note the
   "updated Ns ago" hint next to the refresh button.
2. **Apps** — deploy `go-api`; the Deploy button stays busy until the job actually
   finishes (not a fixed delay), then a success/failure toast fires and the row
   refreshes. Destroy asks for confirmation. The Logs button warns if the logging
   component (Loki) isn't installed instead of opening a dead Grafana link.
3. **Scenarios** — start `observability-sre`; progress streams live; the Activate
   button stays busy until completion, then it shows as active. Deactivate asks
   for confirmation. With >3 scenarios a filter box appears.
4. **Platform** — every provider per category is listed (e.g. ingress → traefik
   *and* nginx) with its installed state. Installing a provider while a sibling
   is installed asks "Swap provider?" first. Removals ask for confirmation.

## Exercise the edge cases

1. **Server down** — stop `labctl ui` while the page is open: a yellow
   "Live updates unavailable — reconnecting…" banner appears and the connection
   dot turns red. Restart the server: the UI reconnects (exponential backoff),
   toasts "Reconnected", and refreshes data.
2. **Job recovery** — start a slow action (e.g. scenario up), reload the page
   mid-run: `GET /api/jobs` shows it `running`; once it ends the views refresh on
   the action_end event. After a disconnect, tracked jobs are settled from
   `/api/jobs` on reconnect.
3. **Cluster down** — with no cluster (`labctl runtime down`), the Dashboard
   shows a "Cluster is unreachable" warning banner, not a blank page.
4. **Cross-site protection** — `curl -X POST -H "Origin: https://evil.example.com"
   localhost:3939/api/apps/go-api/deploy` returns **403**; without an Origin
   header (CLI) it returns **202**.
5. **Double-submit** — clicking Deploy twice fast only fires one job (button
   disables until the job ends).

## Expected

- `bin/labctl ui` serves the SPA with **no external files** (move/rename `ui/dist`
  and it still loads — proves embedding).
- Long actions stream live over WebSocket (depends on Task 003) and survive idle
  periods (server pings every 25 s).
- `GET /api/jobs` returns the recent action history (newest first) with
  `running|succeeded|failed` status.
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
| POST returns 403 `forbidden_origin` | Browser sent a non-local Origin. Access the UI via the same host as the server (or localhost); only same-host / localhost origins may trigger actions. |
| Button stuck on "Working…" | Job end event lost — it self-clears after 15 min; check `GET /api/jobs` and the output panel. |
