import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { PlatformStatus } from '../types'
import { Badge } from '../components/Badge'

interface PlatformProps {
  notify: (level: 'info' | 'success' | 'error', title: string, detail?: string) => void
}

const CATEGORIES = ['ingress', 'metrics', 'logging', 'tracing', 'gitops', 'chaos', 'policy', 'secrets']

export function Platform({ notify }: PlatformProps) {
  const [platform, setPlatform] = useState<PlatformStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<Record<string, boolean>>({})

  const load = useCallback(async () => {
    // /api/platform may not exist separately from /api/status; fall back to status
    try {
      const p = await api.getPlatform()
      setPlatform(p)
    } catch {
      try {
        const s = await api.getStatus()
        setPlatform(s.platform)
      } catch (e) {
        notify('error', 'Failed to load platform status', String(e))
      }
    } finally {
      setLoading(false)
    }
  }, [notify])

  useEffect(() => { load() }, [load])

  async function install(cat: string, name: string) {
    const key = `${cat}/${name}`
    setBusy(b => ({ ...b, [key]: true }))
    try {
      await api.componentUp(cat, name)
      notify('info', `Installing ${name}`, `Category: ${cat}`)
      setTimeout(load, 1500)
    } catch (e) {
      notify('error', `Install ${name} failed`, String(e))
    } finally {
      setBusy(b => ({ ...b, [key]: false }))
    }
  }

  async function remove(cat: string, name: string) {
    const key = `${cat}/${name}`
    setBusy(b => ({ ...b, [key]: true }))
    try {
      await api.componentDown(cat, name)
      notify('info', `Removing ${name}`, `Category: ${cat}`)
      setTimeout(load, 1500)
    } catch (e) {
      notify('error', `Remove ${name} failed`, String(e))
    } finally {
      setBusy(b => ({ ...b, [key]: false }))
    }
  }

  async function installAll() {
    try {
      await api.platformUp()
      notify('info', 'Platform install started', 'Installing all configured platform components…')
      setTimeout(load, 3000)
    } catch (e) {
      notify('error', 'Platform up failed', String(e))
    }
  }

  async function removeAll() {
    try {
      await api.platformDown()
      notify('info', 'Platform removal started', 'Removing all platform components…')
      setTimeout(load, 3000)
    } catch (e) {
      notify('error', 'Platform down failed', String(e))
    }
  }

  if (loading) return <div className="loading">Loading platform status…</div>

  const present = platform
    ? CATEGORIES.filter(k => platform[k])
    : []

  return (
    <div className="card">
      <div className="card-header">
        <span className="card-title">Platform Components</span>
        <button className="btn btn-sm" onClick={load}>Refresh</button>
      </div>

      {present.length === 0 ? (
        <div className="empty-state">No platform data available</div>
      ) : (
        present.map(cat => {
          const comp = platform![cat]!
          const key = `${cat}/${comp.provider}`
          return (
            <div key={cat} className="platform-row">
              <div>
                <div className="platform-name" style={{ textTransform: 'capitalize' }}>{cat}</div>
                {comp.provider && (
                  <div style={{ color: 'var(--muted)', fontSize: 13 }}>{comp.provider}</div>
                )}
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                <Badge variant={comp.active ? 'running' : 'stopped'}>
                  {comp.active ? 'Running' : 'Stopped'}
                </Badge>
                {comp.provider && (
                  comp.active ? (
                    <button
                      className="btn btn-sm btn-danger"
                      disabled={busy[key]}
                      onClick={() => remove(cat, comp.provider)}
                    >
                      {busy[key] ? 'Removing…' : 'Remove'}
                    </button>
                  ) : (
                    <button
                      className="btn btn-sm btn-primary"
                      disabled={busy[key]}
                      onClick={() => install(cat, comp.provider)}
                    >
                      {busy[key] ? 'Installing…' : 'Install'}
                    </button>
                  )
                )}
              </div>
            </div>
          )
        })
      )}

      <div className="card-footer">
        <button className="btn btn-primary" onClick={installAll}>Install All</button>
        <button className="btn btn-danger" onClick={removeAll}>Remove All</button>
      </div>
    </div>
  )
}
