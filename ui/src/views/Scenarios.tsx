import { useState, useEffect, useCallback } from 'react'
import { api } from '../api/client'
import type { Scenario } from '../types'
import { Badge } from '../components/Badge'

interface ScenariosProps {
  notify: (level: 'info' | 'success' | 'error', title: string, detail?: string) => void
}

export function Scenarios({ notify }: ScenariosProps) {
  const [scenarios, setScenarios] = useState<Scenario[]>([])
  const [loading, setLoading] = useState(true)
  const [detail, setDetail] = useState<Scenario | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [busy, setBusy] = useState<Record<string, boolean>>({})

  const load = useCallback(async () => {
    try {
      const list = await api.listScenarios()
      setScenarios(list)
    } catch (e) {
      notify('error', 'Failed to load scenarios', String(e))
    } finally {
      setLoading(false)
    }
  }, [notify])

  useEffect(() => { load() }, [load])

  async function openDetail(name: string) {
    setDetailLoading(true)
    setDetail({ name, displayName: name, description: '', category: '', active: false } as Scenario)
    try {
      const s = await api.getScenario(name)
      setDetail(s)
    } catch (e) {
      notify('error', 'Failed to load scenario details', String(e))
      setDetail(null)
    } finally {
      setDetailLoading(false)
    }
  }

  async function activate(name: string) {
    setBusy(b => ({ ...b, [name]: true }))
    try {
      await api.scenarioUp(name)
      notify('info', `${name} activating`, 'Deploying scenario components…')
      setTimeout(load, 2000)
    } catch (e) {
      notify('error', `Activate ${name} failed`, String(e))
    } finally {
      setBusy(b => ({ ...b, [name]: false }))
    }
  }

  async function deactivate(name: string) {
    setBusy(b => ({ ...b, [name]: true }))
    try {
      await api.scenarioDown(name)
      notify('info', `${name} deactivating`, 'Removing scenario components…')
      setTimeout(load, 2000)
    } catch (e) {
      notify('error', `Deactivate ${name} failed`, String(e))
    } finally {
      setBusy(b => ({ ...b, [name]: false }))
    }
  }

  function copyCmd(cmd: string) {
    navigator.clipboard?.writeText(cmd).catch(() => {
      const ta = document.createElement('textarea')
      ta.value = cmd
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
    })
    notify('success', 'Copied!', '')
  }

  if (loading) return <div className="loading">Loading scenarios…</div>

  return (
    <>
      <div className="card">
        <div className="card-header">
          <span className="card-title">Scenarios ({scenarios.length})</span>
          <button className="btn btn-sm" onClick={load}>Refresh</button>
        </div>

        {scenarios.length === 0 ? (
          <div className="empty-state">No scenarios found</div>
        ) : (
          scenarios.map(s => (
            <div key={s.name} className="scenario-row">
              <div className="scenario-info">
                <div className="scenario-name">{s.displayName || s.name}</div>
                {s.description && (
                  <div className="scenario-desc">
                    {s.description.length > 120 ? s.description.slice(0, 120) + '…' : s.description}
                  </div>
                )}
                <div className="scenario-tags">
                  {s.category && <Badge variant="category">{s.category}</Badge>}
                  {(s.runtimes || []).map(r => (
                    <Badge key={r} variant="runtime">{r}</Badge>
                  ))}
                </div>
              </div>
              <Badge variant={s.active ? 'running' : 'stopped'}>{s.active ? 'Active' : 'Inactive'}</Badge>
              <div className="scenario-actions">
                <button
                  className="btn btn-sm btn-primary"
                  disabled={s.active || busy[s.name]}
                  onClick={() => activate(s.name)}
                >
                  Activate
                </button>
                <button
                  className="btn btn-sm btn-danger"
                  disabled={!s.active || busy[s.name]}
                  onClick={() => deactivate(s.name)}
                >
                  Deactivate
                </button>
                <button className="btn btn-sm" onClick={() => openDetail(s.name)}>Details</button>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Detail modal */}
      {detail && (
        <div className="modal-overlay" onClick={e => { if (e.target === e.currentTarget) setDetail(null) }}>
          <div className="modal-card">
            <button className="modal-close" onClick={() => setDetail(null)}>×</button>

            {detailLoading ? (
              <div className="loading">Loading…</div>
            ) : (
              <>
                <div className="modal-header">
                  <h2>{detail.displayName || detail.name}</h2>
                  <div className="modal-meta">
                    {detail.category && <Badge variant="category">{detail.category}</Badge>}
                    <Badge variant={detail.active ? 'running' : 'stopped'}>{detail.active ? 'Active' : 'Inactive'}</Badge>
                  </div>
                </div>

                {detail.description && (
                  <div className="modal-section">
                    <h3>Description</h3>
                    <p>{detail.description}</p>
                  </div>
                )}

                {detail.prerequisites && (
                  <div className="modal-section">
                    <h3>Prerequisites</h3>
                    <div className="prereq-chips">
                      {(detail.prerequisites.platform || []).map(p => (
                        <Badge key={p} variant="category">Platform: {p}</Badge>
                      ))}
                      {(detail.prerequisites.apps || []).map(a => (
                        <Badge key={a} variant="category">App: {a}</Badge>
                      ))}
                    </div>
                  </div>
                )}

                {detail.components && detail.components.length > 0 && (
                  <div className="modal-section">
                    <h3>Components</h3>
                    <table className="modal-table">
                      <thead>
                        <tr><th>Name</th><th>Type</th><th>Namespace</th><th>Details</th></tr>
                      </thead>
                      <tbody>
                        {detail.components.map(c => (
                          <tr key={c.name}>
                            <td>{c.name}</td>
                            <td><Badge variant="category">{c.type}</Badge></td>
                            <td style={{ color: 'var(--muted)' }}>{c.namespace || 'default'}</td>
                            <td style={{ color: 'var(--muted)', fontSize: 12 }}>{c.chart || c.path || c.script || ''}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}

                {detail.explore?.urls && detail.explore.urls.length > 0 && (
                  <div className="modal-section">
                    <h3>Explore URLs</h3>
                    {detail.explore.urls.map(u => (
                      <div key={u.label} style={{ marginBottom: 6 }}>
                        <a href={u.url} target="_blank" rel="noopener noreferrer">{u.label}</a>
                        <span style={{ color: 'var(--muted)', fontSize: 12, marginLeft: 8 }}>{u.url}</span>
                      </div>
                    ))}
                  </div>
                )}

                {detail.explore?.commands && detail.explore.commands.length > 0 && (
                  <div className="modal-section">
                    <h3>Explore Commands</h3>
                    {detail.explore.commands.map(c => (
                      <div key={c.label} className="cmd-block">
                        <div className="cmd-label">{c.label}</div>
                        <code>{c.command}</code>
                        <button className="cmd-copy" onClick={() => copyCmd(c.command)}>Copy</button>
                      </div>
                    ))}
                  </div>
                )}

                {detail.explore?.tips && detail.explore.tips.length > 0 && (
                  <div className="modal-section">
                    <h3>Tips</h3>
                    <ul>{detail.explore.tips.map((t, i) => <li key={i}>{t}</li>)}</ul>
                  </div>
                )}

                <div className="card-footer">
                  <button
                    className="btn btn-primary"
                    disabled={detail.active || busy[detail.name]}
                    onClick={() => { activate(detail.name); setDetail(null) }}
                  >
                    Activate
                  </button>
                  <button
                    className="btn btn-danger"
                    disabled={!detail.active || busy[detail.name]}
                    onClick={() => { deactivate(detail.name); setDetail(null) }}
                  >
                    Deactivate
                  </button>
                  <button className="btn" onClick={() => setDetail(null)}>Close</button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </>
  )
}
