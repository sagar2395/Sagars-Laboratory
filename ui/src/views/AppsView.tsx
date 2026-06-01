import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { AppStatusResp } from '../api/client';
import { Package, UploadCloud, Trash2 } from 'lucide-react';

export function AppsView() {
  const [apps, setApps] = useState<AppStatusResp[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchApps = () => {
    fetch('/api/apps')
      .then(res => res.json())
      .then(setApps)
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

  const handleDestroy = async (name: string) => {
    await apiClient.appDestroy(name);
  };

  if (loading) return <div className="p-8">Loading...</div>;

  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-6 animate-fade-in">
      <h2 className="text-2xl font-bold">Applications</h2>
      
      <div className="grid grid-cols-1 gap-4">
        {apps.map(app => (
          <div key={app.name} className="surface p-5 flex items-center justify-between">
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
                    onClick={() => handleDestroy(app.name)}
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
    </div>
  );
}
