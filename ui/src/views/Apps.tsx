import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { AppInfo, StatusResponse } from '../types'
import { DeployedBadge } from '../components/Badge'

interface AppsProps {
  notify: (level: 'info' | 'success' | 'error', title: string, detail?: string) => void
  domainSuffix?: string
}

export function Apps({ notify, domainSuffix }: AppsProps) {
  const [apps, setApps] = useState<AppInfo[]>([])
  const [suffix, setSuffix] = useState(domainSuffix ?? '')
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<Record<string, boolean>>({})

  const load = useCallback(async () => {
    try {
      // Use /status to get apps + domainSuffix in one call
      const s: StatusResponse = await api.getStatus()
      setApps(s.apps ?? [])
      if (s.domainSuffix) setSuffix(s.domainSuffix)
    } catch (e) {
      notify('error', 'Failed to load apps', String(e))
    } finally {
      setLoading(false)
    }
  }, [notify])

  useEffect(() => { load() }, [load])

  async function act(
    name: string,
    label: string,
    fn: (name: string) => Promise<unknown>
  ) {
    setBusy(b => ({ ...b, [name]: true }))
    try {
      await fn(name)
      notify('info', label, `${name}`)
      setTimeout(load, 2000)
    } catch (e) {
      notify('error', `${label} failed`, String(e))
    } finally {
      setBusy(b => ({ ...b, [name]: false }))
    }
  }

  function openLogs(name: string) {
    if (!suffix) { notify('error', 'Domain suffix not loaded', ''); return }
    const query = `{namespace="${name}"}`
    const explore = JSON.stringify({ datasource: 'Loki', queries: [{ refId: 'A', expr: query }] })
    window.open(`http://grafana.${suffix}/explore?orgId=1&left=${encodeURIComponent(explore)}`, '_blank')
  }

  if (loading) return <div className="loading">Loading apps…</div>

  return (
    <div className="card">
      <div className="card-header">
        <span className="card-title">Applications ({apps.length})</span>
        <button className="btn btn-sm" onClick={load}>Refresh</button>
      </div>

      {apps.length === 0 ? (
        <div className="empty-state">No applications found</div>
      ) : (
        apps.map(a => (
          <div key={a.name} className="app-row">
            <div style={{ flex: 1, minWidth: 120 }}>
              <div className="app-name">{a.name}</div>
              {a.namespace && (
                <div style={{ color: 'var(--muted)', fontSize: 12 }}>ns: {a.namespace}</div>
              )}
            </div>

            <div className="app-meta">
              <span>build: {a.buildStrategy || '—'}</span>
              <span style={{ margin: '0 6px' }}>·</span>
              <span>deploy: {a.deployStrategy || '—'}</span>
              {a.replicas && <><span style={{ margin: '0 6px' }}>·</span><span>replicas: {a.replicas}</span></>}
              {a.ready    && <><span style={{ margin: '0 6px' }}>·</span><span>ready: {a.ready}</span></>}
            </div>

            <DeployedBadge deployed={a.deployed} />

            <div className="btn-group">
              <button
                className="btn btn-sm"
                onClick={() => openLogs(a.name)}
                title="Open logs in Grafana"
              >
                Logs
              </button>
              <button
                className="btn btn-sm"
                disabled={busy[a.name]}
                onClick={() => act(a.name, 'Build started', api.buildApp)}
              >
                Build
              </button>
              <button
                className="btn btn-sm btn-primary"
                disabled={busy[a.name]}
                onClick={() => act(a.name, 'Deploy started', api.deployApp)}
              >
                Deploy
              </button>
              <button
                className="btn btn-sm btn-danger"
                disabled={busy[a.name]}
                onClick={() => act(a.name, 'Destroy started', api.destroyApp)}
              >
                Destroy
              </button>
            </div>
          </div>
        ))
      )}
    </div>
  )
}
