import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { StatusResponse, DashboardURL } from '../types'
import { Badge } from '../components/Badge'

const REFRESH_SVG = (
  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <polyline points="23 4 23 10 17 10" />
    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
  </svg>
)

interface DashboardProps {
  notify: (level: 'info' | 'success' | 'error', title: string, detail?: string) => void
}

export function Dashboard({ notify }: DashboardProps) {
  const [status, setStatus] = useState<StatusResponse | null>(null)
  const [dashboards, setDashboards] = useState<DashboardURL[]>([])
  const [loading, setLoading] = useState(true)
  const [refreshing, setRefreshing] = useState(false)

  const load = useCallback(async () => {
    try {
      const [s, d] = await Promise.all([
        api.getStatus(),
        api.getDashboards().catch(() => [] as DashboardURL[]),
      ])
      setStatus(s)
      setDashboards(d)
    } catch (e) {
      notify('error', 'Failed to load status', String(e))
    } finally {
      setLoading(false)
      setRefreshing(false)
    }
  }, [notify])

  useEffect(() => { load() }, [load])

  // Reload every 30 s as backup
  useEffect(() => {
    const t = setInterval(load, 30_000)
    return () => clearInterval(t)
  }, [load])

  async function handleRefresh() {
    setRefreshing(true)
    await load()
  }

  async function platformUp() {
    try {
      await api.platformUp()
      notify('info', 'Platform install started', 'Installing all platform components…')
    } catch (e) { notify('error', 'Platform up failed', String(e)) }
  }

  async function platformDown() {
    try {
      await api.platformDown()
      notify('info', 'Platform removal started', 'Removing platform components…')
    } catch (e) { notify('error', 'Platform down failed', String(e)) }
  }

  async function componentUp(cat: string, name: string) {
    try {
      await api.componentUp(cat, name)
      notify('info', `Installing ${name}`, '')
    } catch (e) { notify('error', `Install ${name} failed`, String(e)) }
  }

  async function componentDown(cat: string, name: string) {
    try {
      await api.componentDown(cat, name)
      notify('info', `Removing ${name}`, '')
    } catch (e) { notify('error', `Remove ${name} failed`, String(e)) }
  }

  async function deployApp(name: string) {
    try {
      await api.deployApp(name)
      notify('info', `Deploy started`, `Deploying ${name}…`)
    } catch (e) { notify('error', `Deploy ${name} failed`, String(e)) }
  }

  async function destroyApp(name: string) {
    try {
      await api.destroyApp(name)
      notify('info', `Destroy started`, `Destroying ${name}…`)
    } catch (e) { notify('error', `Destroy ${name} failed`, String(e)) }
  }

  if (loading) {
    return <div className="loading">Loading dashboard…</div>
  }

  const c = status?.cluster
  const p = status?.platform ?? {}
  const apps = status?.apps ?? []
  const links = dashboards.filter(d => d.available)
  const platformCategories = ['ingress', 'metrics', 'logging', 'tracing', 'gitops', 'chaos', 'policy', 'secrets']

  return (
    <>
      {/* Quick links */}
      {links.length > 0 && (
        <div className="quick-links">
          {links.map(d => (
            <a key={d.name} className="quick-link" href={d.url} target="_blank" rel="noopener noreferrer">
              {d.label} ↗
            </a>
          ))}
        </div>
      )}

      <div className="grid">
        {/* Cluster card */}
        <div className="card">
          <div className="card-header">
            <span className="card-title">Cluster</span>
            <button className={`btn-icon${refreshing ? ' spinning' : ''}`} onClick={handleRefresh} title="Refresh">
              {REFRESH_SVG}
            </button>
          </div>
          {c ? (
            <div>
              <div className="row"><span className="label">Context</span><span className="value">{c.context || 'N/A'}</span></div>
              <div className="row"><span className="label">K8s Version</span><span className="value">{c.k8sVersion || 'N/A'}</span></div>
              <div className="row"><span className="label">Nodes</span><span className="value">{c.nodeCount}</span></div>
              <div className="row">
                <span className="label">Status</span>
                <Badge variant={c.connected ? 'running' : 'stopped'}>{c.connected ? 'Connected' : 'Disconnected'}</Badge>
              </div>
            </div>
          ) : (
            <div className="empty-state">No cluster data</div>
          )}
        </div>

        {/* Platform components card */}
        <div className="card full-width">
          <div className="card-header">
            <span className="card-title">Platform Components</span>
          </div>
          {platformCategories
            .filter(key => p[key])
            .map(key => {
              const comp = p[key]!
              return (
                <div key={key} className="platform-row">
                  <span className="platform-name">{key}</span>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8, flexWrap: 'wrap' }}>
                    <span className="value">{comp.provider || 'none'}</span>
                    <Badge variant={comp.active ? 'running' : 'stopped'}>{comp.active ? 'Running' : 'Stopped'}</Badge>
                    {comp.provider && (
                      comp.active
                        ? <button className="btn btn-sm btn-danger" onClick={() => componentDown(key, comp.provider)}>Remove</button>
                        : <button className="btn btn-sm btn-primary" onClick={() => componentUp(key, comp.provider)}>Install</button>
                    )}
                  </div>
                </div>
              )
            })}
          {Object.keys(p).length === 0 && <div className="empty-state">No platform data</div>}
          <div className="card-footer">
            <button className="btn btn-primary" onClick={platformUp}>Install Platform</button>
            <button className="btn btn-danger" onClick={platformDown}>Remove Platform</button>
          </div>
        </div>

        {/* Applications card */}
        <div className="card full-width">
          <div className="card-header">
            <span className="card-title">Applications</span>
          </div>
          {apps.length === 0 ? (
            <div className="empty-state">No applications found</div>
          ) : apps.map(a => (
            <div key={a.name} className="app-row">
              <span className="app-name">{a.name}</span>
              <span className="app-meta">
                build: {a.buildStrategy || 'N/A'} · deploy: {a.deployStrategy || 'N/A'}
                {a.replicas && ` · replicas: ${a.replicas}`}
                {a.ready && ` · ready: ${a.ready}`}
              </span>
              <Badge variant={a.deployed ? 'running' : 'stopped'}>{a.deployed ? 'Deployed' : 'Not Deployed'}</Badge>
              <div className="btn-group">
                <button className="btn btn-sm btn-primary" onClick={() => deployApp(a.name)}>Deploy</button>
                <button className="btn btn-sm btn-danger" onClick={() => destroyApp(a.name)}>Destroy</button>
              </div>
            </div>
          ))}
        </div>
      </div>
    </>
  )
}
