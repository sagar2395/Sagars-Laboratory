import { useState, useCallback, useRef } from 'react'
import { useWebSocket, type WSStatus } from './hooks/useWebSocket'
import type { ActionEvent, LogEntry, Notification, NotifLevel } from './types'
import { NotificationList } from './components/Notification'
import { LogPanel } from './components/LogPanel'
import { Dashboard } from './views/Dashboard'
import { Scenarios } from './views/Scenarios'
import { Platform } from './views/Platform'
import { Apps } from './views/Apps'
import { api } from './api/client'

type Tab = 'dashboard' | 'scenarios' | 'platform' | 'apps'

const TABS: { id: Tab; label: string }[] = [
  { id: 'dashboard',  label: 'Dashboard'  },
  { id: 'scenarios',  label: 'Scenarios'  },
  { id: 'platform',   label: 'Platform'   },
  { id: 'apps',       label: 'Apps'       },
]

let notifSeq = 0
let logSeq = 0

function nowHMS() {
  return new Date().toLocaleTimeString('en-US', { hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export default function App() {
  const [tab, setTab] = useState<Tab>('dashboard')
  const [wsStatus, setWsStatus] = useState<WSStatus>('connecting')
  const [notifications, setNotifications] = useState<Notification[]>([])
  const [logEntries, setLogEntries] = useState<LogEntry[]>([])
  const [runtimes, setRuntimes] = useState<string[]>([])
  const [activeRuntime, setActiveRuntime] = useState('')

  // Memoised so WebSocket hook doesn't reconnect on every render
  const onStatusChange = useCallback((s: WSStatus) => setWsStatus(s), [])

  const appendLog = useCallback((level: LogEntry['level'], text: string) => {
    setLogEntries(prev => {
      const next = [...prev, { id: ++logSeq, ts: nowHMS(), level, text }]
      return next.length > 500 ? next.slice(-500) : next
    })
  }, [])

  const onActionEvent = useCallback((ev: ActionEvent) => {
    switch (ev.type) {
      case 'action_start':
        appendLog('cmd', `▸ ${ev.action ?? ''}${ev.command ? `  (${ev.command})` : ''}`)
        break
      case 'action_output':
        appendLog(ev.stream === 'stderr' ? 'stderr' : 'output', ev.output ?? '')
        break
      case 'action_end': {
        const ok = (ev.exitCode ?? -1) === 0
        appendLog(ok ? 'success' : 'error', ok
          ? `✓ ${ev.action ?? 'Command'} completed`
          : `✗ ${ev.action ?? 'Command'} failed${ev.error ? ': ' + ev.error : ''}`)
        break
      }
      case 'action_error':
        appendLog('error', ev.error ?? 'Unknown error')
        break
    }
  }, [appendLog])

  useWebSocket({ onActionEvent, onStatusChange })

  // Load runtimes once
  const runtimeLoaded = useRef(false)
  if (!runtimeLoaded.current) {
    runtimeLoaded.current = true
    api.listRuntimes().then(list => {
      const names = list.map(r => r.name ?? r.profile ?? '').filter(Boolean)
      setRuntimes(names)
      const active = list.find(r => r.active)
      if (active) setActiveRuntime(active.name ?? active.profile ?? '')
    }).catch(() => {})
  }

  const notify = useCallback((level: NotifLevel, title: string, detail?: string) => {
    const id = ++notifSeq
    setNotifications(prev => [...prev, { id, level, title, detail }])
    if (level !== 'error') setTimeout(() => dismiss(id), 8000)
  }, [])

  function dismiss(id: number) {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  async function handleRuntimeChange(value: string) {
    if (!value) return
    try {
      await api.activateRuntime(value)
      setActiveRuntime(value)
      notify('info', 'Runtime switched', value)
    } catch { /* non-critical */ }
  }

  const connDot = wsStatus === 'connected' ? 'green' : wsStatus === 'connecting' ? 'yellow' : 'red'
  const connLabel = wsStatus === 'connected' ? 'Connected' : wsStatus === 'connecting' ? 'Connecting…' : 'Disconnected'

  return (
    <div className="app-shell">
      <NotificationList items={notifications} onDismiss={dismiss} />

      {/* Header */}
      <header className="app-header">
        <h1><span className="accent">labctl</span> Dashboard</h1>
        <div className="header-right">
          {runtimes.length > 0 && (
            <select
              className="runtime-select"
              value={activeRuntime}
              onChange={e => handleRuntimeChange(e.target.value)}
            >
              {runtimes.map(r => (
                <option key={r} value={r}>{r}{r === activeRuntime ? ' (active)' : ''}</option>
              ))}
            </select>
          )}
          <div className="conn-indicator">
            <span className={`dot dot-${connDot}`} />
            <span>{connLabel}</span>
          </div>
        </div>
      </header>

      {/* Nav */}
      <nav className="nav-tabs">
        {TABS.map(t => (
          <button
            key={t.id}
            className={`nav-tab${tab === t.id ? ' active' : ''}`}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </nav>

      {/* Content */}
      <main className="main-content">
        {tab === 'dashboard'  && <Dashboard notify={notify} />}
        {tab === 'scenarios'  && <Scenarios notify={notify} />}
        {tab === 'platform'   && <Platform  notify={notify} />}
        {tab === 'apps'       && <Apps      notify={notify} />}
      </main>

      {/* Log panel */}
      <LogPanel entries={logEntries} onClear={() => setLogEntries([])} />
    </div>
  )
}
