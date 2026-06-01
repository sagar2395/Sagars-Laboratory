import { useEffect, useState } from 'react';
import { Power, RefreshCcw } from 'lucide-react';

export function PlatformView() {
  const [platform, setPlatform] = useState<Record<string, any[]>>({});
  const [loading, setLoading] = useState(true);

  const fetchPlatform = () => {
    fetch('/api/platform/status')
      .then(res => res.json())
      .then(setPlatform)
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchPlatform();
    const int = setInterval(fetchPlatform, 5000);
    return () => clearInterval(int);
  }, []);

  const handleComponentAction = async (category: string, name: string, active: boolean) => {
    if (active) {
      await fetch(`/api/platform/${category}/${name}/down`, { method: 'POST' });
    } else {
      await fetch(`/api/platform/${category}/${name}/up`, { method: 'POST' });
    }
    // Optimistic update
    setPlatform(prev => {
      const next = { ...prev };
      next[category] = next[category].map(p => p.name === name ? { ...p, installed: !active } : p);
      return next;
    });
  };

  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-8 animate-fade-in">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-bold">Platform Components</h2>
        <button className="btn btn-secondary text-sm" onClick={fetchPlatform}>
          <RefreshCcw size={14} className="mr-2" /> Refresh
        </button>
      </div>

      <div className="grid grid-cols-2 gap-8">
        {Object.entries(platform).map(([category, providers]) => (
          <div key={category} className="surface p-6">
            <h3 className="text-lg font-semibold capitalize mb-4 border-b border-[var(--border-light)] pb-2">{category}</h3>
            <div className="flex flex-col gap-3">
              {providers.map(p => (
                <div key={p.name} className={`flex items-center justify-between p-3 rounded-[var(--radius-md)] border ${p.installed ? 'border-[var(--status-success)] bg-[var(--status-success-bg)]' : 'border-[var(--border-light)] bg-[var(--bg-base)]'}`}>
                  <div className="font-medium">{p.name}</div>
                  <button 
                    onClick={() => handleComponentAction(category, p.name, p.installed)}
                    className={`btn px-3 py-1 text-xs ${p.installed ? 'btn-secondary text-[var(--status-danger)]' : 'bg-[var(--bg-surface-elevated)] text-[var(--text-primary)] hover:bg-[var(--accent-primary)] hover:text-white'}`}
                  >
                    <Power size={14} className="mr-1" />
                    {p.installed ? 'Uninstall' : 'Install'}
                  </button>
                </div>
              ))}
              {providers.length === 0 && <div className="text-sm text-[var(--text-muted)]">No providers available.</div>}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
