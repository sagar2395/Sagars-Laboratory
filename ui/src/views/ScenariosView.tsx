<<<<<<< HEAD
import { useEffect, useState, useRef } from 'react';
=======
import { useEffect, useState } from 'react';
>>>>>>> 9b97903 (Fixing conflict)
import { apiClient } from '../api/client';
import type { ScenarioStatus } from '../api/client';
import { useActionEvents } from '../api/useActionEvents';
import { Play, Square, Terminal } from 'lucide-react';

export function ScenariosView() {
  const [scenarios, setScenarios] = useState<ScenarioStatus[]>([]);
  const { events, clearEvents } = useActionEvents();
  const [loading, setLoading] = useState(true);
<<<<<<< HEAD
  const [error, setError] = useState<string | null>(null);
  const terminalEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    terminalEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [events]);

  const fetchScenarios = () => {
    apiClient.getScenarios()
      .then(data => {
        setScenarios(data);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
=======

  const fetchScenarios = () => {
    apiClient.getScenarios().then(setScenarios).finally(() => setLoading(false));
>>>>>>> 9b97903 (Fixing conflict)
  };

  useEffect(() => {
    fetchScenarios();
    // Poll to keep active state synced if someone changes it via CLI
    const int = setInterval(fetchScenarios, 5000);
    return () => clearInterval(int);
  }, []);

  const handleAction = async (name: string, active: boolean) => {
    clearEvents();
    if (active) {
      await apiClient.scenarioDown(name);
    } else {
      await apiClient.scenarioUp(name);
    }
    // Optimistic update
    setScenarios(s => s.map(x => x.name === name ? { ...x, active: !active } : x));
  };

<<<<<<< HEAD
  if (loading && scenarios.length === 0) return <div className="p-8 text-[var(--text-secondary)] animate-pulse">Loading...</div>;
  if (error && scenarios.length === 0) return (
    <div className="p-8 max-w-6xl mx-auto animate-fade-in">
      <div className="surface p-8 border-l-4 border-l-[var(--status-danger)] bg-[var(--status-danger-bg)]">
        <h2 className="text-xl font-bold text-[var(--status-danger)] mb-2">Backend Offline</h2>
        <p className="text-[var(--text-primary)]">The labctl backend could not be reached.</p>
        <p className="text-sm text-[var(--text-secondary)] mt-2">Error: {error}</p>
      </div>
    </div>
  );

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-6 animate-fade-in">
      <div className="flex items-center gap-4">
        <h2 className="text-2xl font-bold">Scenarios</h2>
        {error && <span className="badge badge-danger">Offline</span>}
      </div>
=======
  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-6 animate-fade-in">
      <h2 className="text-2xl font-bold">Scenarios</h2>
>>>>>>> 9b97903 (Fixing conflict)
      <p className="text-[var(--text-secondary)] mb-4">Declarative configurations to explore platform engineering concepts.</p>
      
      <div className="grid grid-cols-2 gap-6">
        {/* Scenarios List */}
        <div className="flex flex-col gap-4">
          {scenarios.map(sc => (
<<<<<<< HEAD
            <div key={sc.name} className="surface surface-interactive p-5 flex flex-col gap-3">
=======
            <div key={sc.name} className="surface p-5 flex flex-col gap-3">
>>>>>>> 9b97903 (Fixing conflict)
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-semibold flex items-center gap-2">
                    {sc.name}
<<<<<<< HEAD
                    {sc.active && <span className="status-dot-success ml-2"></span>}
=======
                    {sc.active && <span className="badge badge-success">Active</span>}
>>>>>>> 9b97903 (Fixing conflict)
                  </h3>
                  <span className="text-xs text-[var(--accent-primary)] bg-[var(--accent-primary)] bg-opacity-10 px-2 py-1 rounded mt-1 inline-block">
                    {sc.category}
                  </span>
                </div>
                <button 
                  onClick={() => handleAction(sc.name, sc.active)}
                  className={`btn ${sc.active ? 'btn-secondary text-[var(--status-danger)]' : 'btn-primary'}`}
                >
                  {sc.active ? <Square size={16} className="mr-2"/> : <Play size={16} className="mr-2"/>}
                  {sc.active ? 'Stop' : 'Start'}
                </button>
              </div>
              <p className="text-sm text-[var(--text-secondary)]">{sc.description}</p>
            </div>
          ))}
        </div>

        {/* Real-time Event Stream */}
<<<<<<< HEAD
        <div className="surface-glass p-0 flex flex-col h-[500px] shadow-[inset_0_2px_10px_rgba(0,0,0,0.5)] overflow-hidden">
          <div className="p-3 border-b border-[var(--border-light)] bg-[#1a1b26] flex items-center">
            <div className="flex gap-2 ml-1">
              <div className="w-3 h-3 rounded-full bg-[#ff5f56] shadow-sm"></div>
              <div className="w-3 h-3 rounded-full bg-[#ffbd2e] shadow-sm"></div>
              <div className="w-3 h-3 rounded-full bg-[#27c93f] shadow-sm"></div>
            </div>
            <div className="flex items-center gap-2 mx-auto text-[var(--text-muted)]">
              <Terminal size={14} />
              <span className="text-xs font-semibold uppercase tracking-wider">WebSocket Stream</span>
            </div>
            <div className="w-12"></div> {/* spacer for centering */}
          </div>
          <div className="p-4 flex-1 overflow-auto font-mono text-xs bg-[#0f0f14] flex flex-col gap-1.5 shadow-[inset_0_0_20px_rgba(0,0,0,0.8)]">
=======
        <div className="surface-glass p-0 flex flex-col h-[500px] border border-[var(--border-light)] overflow-hidden">
          <div className="p-3 border-b border-[var(--border-light)] bg-[var(--bg-surface-elevated)] flex items-center gap-2">
            <Terminal size={16} className="text-[var(--text-muted)]" />
            <span className="text-sm font-medium text-[var(--text-secondary)]">Action Event Stream</span>
          </div>
          <div className="p-4 flex-1 overflow-auto font-mono text-xs bg-[#0a0a0f] flex flex-col gap-1">
>>>>>>> 9b97903 (Fixing conflict)
            {events.length === 0 ? (
              <div className="text-[var(--text-muted)] italic">Awaiting actions...</div>
            ) : (
              events.map((e, i) => (
<<<<<<< HEAD
                <div key={i} className="flex gap-3 text-[var(--text-primary)] hover:bg-[rgba(255,255,255,0.02)] p-1.5 rounded transition-colors">
                  <span className="text-[var(--text-muted)] shrink-0 opacity-70">{new Date(e.Timestamp).toLocaleTimeString()}</span>
                  <span className={`${e.Type === 'action_end' && e.ExitCode === 0 ? 'text-[var(--status-success)]' : e.Type === 'action_end' ? 'text-[var(--status-danger)]' : 'text-[var(--accent-primary)]'}`}>
                    [{e.Type}]
                  </span>
                  <span className="break-all font-medium leading-relaxed">{e.Line || e.Action || e.Error}</span>
                </div>
              ))
            )}
            <div ref={terminalEndRef} />
=======
                <div key={i} className="flex gap-2 text-[var(--text-primary)] hover:bg-[#1a1a24] p-1 rounded">
                  <span className="text-[var(--text-muted)] shrink-0">{new Date(e.Timestamp).toLocaleTimeString()}</span>
                  <span className={`${e.Type === 'action_end' && e.ExitCode === 0 ? 'text-[var(--status-success)]' : e.Type === 'action_end' ? 'text-[var(--status-danger)]' : 'text-[var(--accent-primary)]'}`}>
                    [{e.Type}]
                  </span>
                  <span className="break-all">{e.Line || e.Action || e.Error}</span>
                </div>
              ))
            )}
>>>>>>> 9b97903 (Fixing conflict)
          </div>
        </div>
      </div>
    </div>
  );
}
