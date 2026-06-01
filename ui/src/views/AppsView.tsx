import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { AppStatusResp } from '../api/client';
import { Package, UploadCloud, Trash2 } from 'lucide-react';
import { Modal } from '../components/Modal';

export function AppsView() {
  const [apps, setApps] = useState<AppStatusResp[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [deleteConfirmApp, setDeleteConfirmApp] = useState<string | null>(null);

  const fetchApps = () => {
    fetch('/api/apps')
      .then(res => {
        if (!res.ok) throw new Error('Failed to fetch applications');
        return res.json();
      })
      .then(data => {
        setApps(data);
        setError(null);
      })
      .catch(err => setError(err.message))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchApps();
    const int = setInterval(fetchApps, 5000);
    return () => clearInterval(int);
  }, []);

  const handleDeploy = async (name: string) => {
    await apiClient.appDeploy(name);
    // Let polling catch the update
  };

  const confirmDestroy = (name: string) => setDeleteConfirmApp(name);
  const cancelDestroy = () => setDeleteConfirmApp(null);

  const handleDestroy = async () => {
    if (!deleteConfirmApp) return;
    await apiClient.appDestroy(deleteConfirmApp);
    setDeleteConfirmApp(null);
  };

  if (loading && apps.length === 0) return <div className="p-8 text-[var(--text-secondary)] animate-pulse">Loading...</div>;
  if (error && apps.length === 0) return (
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
        <h2 className="text-2xl font-bold">Applications</h2>
        {error && <span className="badge badge-danger">Offline</span>}
      </div>
      
      <div className="grid grid-cols-1 gap-4">
        {apps.map(app => (
          <div key={app.name} className="surface surface-interactive p-5 flex items-center justify-between">
            <div className="flex items-start gap-4">
              <div className="p-3 bg-[var(--bg-surface-elevated)] rounded-full text-[var(--accent-primary)]">
                <Package size={24} />
              </div>
              <div>
                <h3 className="text-lg font-semibold">{app.name}</h3>
                <div className="text-sm text-[var(--text-secondary)] flex gap-4 mt-1">
                  <span>Build: <code className="text-xs bg-[#000] px-1 rounded">{app.buildStrategy}</code></span>
                  <span>Deploy: <code className="text-xs bg-[#000] px-1 rounded">{app.deployStrategy}</code></span>
                </div>
              </div>
            </div>
            
            <div className="flex items-center gap-6">
              {app.deployed ? (
                <div className="flex flex-col items-end">
                   <span className="badge badge-success mb-1">Running</span>
                   <span className="text-xs text-[var(--text-secondary)]">{app.ready}/{app.replicas} Replicas Ready</span>
                </div>
              ) : (
                <span className="badge badge-warning">Not Deployed</span>
              )}
              
              <div className="flex gap-2 border-l border-[var(--border-light)] pl-6">
                <button 
                  onClick={() => handleDeploy(app.name)}
                  className="btn btn-primary px-3 py-1.5 text-sm"
                >
                  <UploadCloud size={16} className="mr-2" /> Deploy
                </button>
                {app.deployed && (
                  <button 
                    onClick={() => confirmDestroy(app.name)}
                    className="btn btn-secondary text-[var(--status-danger)] px-3 py-1.5 text-sm"
                  >
                    <Trash2 size={16} />
                  </button>
                )}
              </div>
            </div>
          </div>
        ))}
        {apps.length === 0 && <div className="surface p-8 text-center text-[var(--text-muted)]">No applications found in the workspace.</div>}
      </div>

      <Modal 
        isOpen={deleteConfirmApp !== null} 
        onClose={cancelDestroy} 
        title="Confirm Deletion"
      >
        <p className="text-[var(--text-secondary)] mb-6 leading-relaxed">
          Are you sure you want to destroy the application <strong className="text-[var(--text-primary)]">{deleteConfirmApp}</strong>? This action will remove all deployed resources from the cluster.
        </p>
        <div className="flex justify-end gap-3 mt-4">
          <button className="btn btn-secondary" onClick={cancelDestroy}>Cancel</button>
          <button className="btn btn-danger" onClick={handleDestroy}>Destroy App</button>
        </div>
      </Modal>
    </div>
  );
}
