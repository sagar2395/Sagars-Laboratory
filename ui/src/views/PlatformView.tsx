import { useEffect, useState } from 'react';
import { Power, RefreshCcw } from 'lucide-react';
import { Modal } from '../components/Modal';

export function PlatformView() {
  const [platform, setPlatform] = useState<Record<string, any[]>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [uninstallConfirm, setUninstallConfirm] = useState<{category: string, name: string} | null>(null);

  const fetchPlatform = () => {
    fetch('/api/platform/status')
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch platform status');
        return res.json();
      })
      .then(data => {
        setPlatform(data);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchPlatform();
    const int = setInterval(fetchPlatform, 5000);
    return () => clearInterval(int);
  }, []);

  const handleComponentAction = async (category: string, name: string, active: boolean) => {
    if (active) {
      setUninstallConfirm({ category, name });
    } else {
      await fetch(`/api/platform/${category}/${name}/up`, { method: 'POST' });
      // Optimistic update
      setPlatform(prev => {
        const next = { ...prev };
        next[category] = next[category].map(p => p.name === name ? { ...p, installed: true } : p);
        return next;
      });
    }
  };

  const confirmUninstall = async () => {
    if (!uninstallConfirm) return;
    const { category, name } = uninstallConfirm;
    setUninstallConfirm(null);
    await fetch(`/api/platform/${category}/${name}/down`, { method: 'POST' });
    setPlatform(prev => {
      const next = { ...prev };
      next[category] = next[category].map(p => p.name === name ? { ...p, installed: false } : p);
      return next;
    });
  };

  if (loading && Object.keys(platform).length === 0) return <div className="p-8 text-[var(--text-secondary)] animate-pulse">Loading...</div>;
  if (error && Object.keys(platform).length === 0) return (
    <div className="p-8 max-w-6xl mx-auto animate-fade-in">
      <div className="surface p-8 border-l-4 border-l-[var(--status-danger)] bg-[var(--status-danger-bg)]">
        <h2 className="text-xl font-bold text-[var(--status-danger)] mb-2">Backend Offline</h2>
        <p className="text-[var(--text-primary)]">The labctl backend could not be reached.</p>
        <p className="text-sm text-[var(--text-secondary)] mt-2">Error: {error}</p>
      </div>
    </div>
  );

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-8 animate-fade-in">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h2 className="text-2xl font-bold">Platform Components</h2>
          {error && <span className="badge badge-danger">Offline</span>}
        </div>
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

      <Modal 
        isOpen={uninstallConfirm !== null} 
        onClose={() => setUninstallConfirm(null)} 
        title="Confirm Uninstall"
      >
        <p className="text-[var(--text-secondary)] mb-6 leading-relaxed">
          Are you sure you want to uninstall <strong className="text-[var(--text-primary)]">{uninstallConfirm?.name}</strong> from the {uninstallConfirm?.category} platform? This may disrupt cluster services.
        </p>
        <div className="flex justify-end gap-3 mt-4">
          <button className="btn btn-secondary" onClick={() => setUninstallConfirm(null)}>Cancel</button>
          <button className="btn btn-danger" onClick={confirmUninstall}>Uninstall</button>
        </div>
      </Modal>
    </div>
  );
}
