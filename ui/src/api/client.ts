import type {
  StatusResponse,
  AppInfo,
  PlatformStatus,
  DashboardURL,
  Scenario,
  Runtime,
} from '../types'

const BASE = '/api'

/** Generic fetch wrapper — throws Error with the server's error message on failures. */
async function req<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, opts)
  if (!res.ok) {
    let msg = `HTTP ${res.status}`
    try {
      const body = await res.json() as { error?: string }
      if (body.error) msg = body.error
    } catch { /* ignore parse error */ }
    throw new Error(msg)
  }
  return res.json() as Promise<T>
}

export const api = {
  // ── Status ──────────────────────────────────────────────────────────────
  getStatus:    ()           => req<StatusResponse>('/status'),
  getDashboards: ()          => req<DashboardURL[]>('/dashboards'),

  // ── Apps ────────────────────────────────────────────────────────────────
  listApps:   ()             => req<AppInfo[]>('/apps'),
  buildApp:   (name: string) => req<unknown>(`/apps/${enc(name)}/build`,   { method: 'POST' }),
  deployApp:  (name: string) => req<unknown>(`/apps/${enc(name)}/deploy`,  { method: 'POST' }),
  destroyApp: (name: string) => req<unknown>(`/apps/${enc(name)}/destroy`, { method: 'POST' }),

  // ── Platform ─────────────────────────────────────────────────────────────
  getPlatform:   ()                          => req<PlatformStatus>('/platform'),
  platformUp:    ()                          => req<unknown>('/platform/up',   { method: 'POST' }),
  platformDown:  ()                          => req<unknown>('/platform/down', { method: 'POST' }),
  componentUp:   (cat: string, name: string) => req<unknown>(`/platform/component/${enc(cat)}/${enc(name)}/up`,   { method: 'POST' }),
  componentDown: (cat: string, name: string) => req<unknown>(`/platform/component/${enc(cat)}/${enc(name)}/down`, { method: 'POST' }),

  // ── Scenarios ─────────────────────────────────────────────────────────────
  listScenarios: ()             => req<Scenario[]>('/scenarios'),
  getScenario:   (name: string) => req<Scenario>(`/scenarios/${enc(name)}`),
  scenarioUp:    (name: string) => req<unknown>(`/scenarios/${enc(name)}/up`,   { method: 'POST' }),
  scenarioDown:  (name: string) => req<unknown>(`/scenarios/${enc(name)}/down`, { method: 'POST' }),

  // ── Runtimes ──────────────────────────────────────────────────────────────
  listRuntimes:     ()             => req<Runtime[]>('/runtimes'),
  activateRuntime:  (name: string) => req<unknown>(`/runtimes/${enc(name)}/activate`,   { method: 'POST' }),
  deactivateRuntime:(name: string) => req<unknown>(`/runtimes/${enc(name)}/deactivate`, { method: 'POST' }),
}

function enc(s: string) { return encodeURIComponent(s) }
