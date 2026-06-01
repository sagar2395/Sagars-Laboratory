import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { ScenarioStatus } from '../api/client';
import { useActionEvents } from '../api/useActionEvents';
import { Play, Square, Terminal } from 'lucide-react';

export function ScenariosView() {
  const [scenarios, setScenarios] = useState<ScenarioStatus[]>([]);
  const { events, clearEvents } = useActionEvents();
  const [loading, setLoading] = useState(true);

  const fetchScenarios = () => {
    apiClient.getScenarios().then(setScenarios).finally(() => setLoading(false));
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

  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-6 animate-fade-in">
      <h2 className="text-2xl font-bold">Scenarios</h2>
      <p className="text-[var(--text-secondary)] mb-4">Declarative configurations to explore platform engineering concepts.</p>
      
      <div className="grid grid-cols-2 gap-6">
        {/* Scenarios List */}
        <div className="flex flex-col gap-4">
          {scenarios.map(sc => (
            <div key={sc.name} className="surface p-5 flex flex-col gap-3">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-semibold flex items-center gap-2">
                    {sc.name}
                    {sc.active && <span className="badge badge-success">Active</span>}
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
        <div className="surface-glass p-0 flex flex-col h-[500px] border border-[var(--border-light)] overflow-hidden">
          <div className="p-3 border-b border-[var(--border-light)] bg-[var(--bg-surface-elevated)] flex items-center gap-2">
            <Terminal size={16} className="text-[var(--text-muted)]" />
            <span className="text-sm font-medium text-[var(--text-secondary)]">Action Event Stream</span>
          </div>
          <div className="p-4 flex-1 overflow-auto font-mono text-xs bg-[#0a0a0f] flex flex-col gap-1">
            {events.length === 0 ? (
              <div className="text-[var(--text-muted)] italic">Awaiting actions...</div>
            ) : (
              events.map((e, i) => (
                <div key={i} className="flex gap-2 text-[var(--text-primary)] hover:bg-[#1a1a24] p-1 rounded">
                  <span className="text-[var(--text-muted)] shrink-0">{new Date(e.Timestamp).toLocaleTimeString()}</span>
                  <span className={`${e.Type === 'action_end' && e.ExitCode === 0 ? 'text-[var(--status-success)]' : e.Type === 'action_end' ? 'text-[var(--status-danger)]' : 'text-[var(--accent-primary)]'}`}>
                    [{e.Type}]
                  </span>
                  <span className="break-all">{e.Line || e.Action || e.Error}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
