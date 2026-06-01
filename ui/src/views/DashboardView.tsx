import { useEffect, useState } from 'react';
import { apiClient } from '../api/client';
import type { StatusResponse } from '../api/client';
import { Server, Activity, Cpu, Package } from 'lucide-react';

export function DashboardView() {
  const [status, setStatus] = useState<StatusResponse | null>(null);
  const [loading, setLoading] = useState(true);
<<<<<<< HEAD
  const [error, setError] = useState<string | null>(null);

  const fetchStatus = () => {
    apiClient.getStatus()
      .then((data) => {
        setStatus(data);
        setError(null);
      })
      .catch((err) => setError(err.message))
      .finally(() => setLoading(false));
=======

  const fetchStatus = () => {
    apiClient.getStatus().then(setStatus).finally(() => setLoading(false));
>>>>>>> 9b97903 (Fixing conflict)
  };

  useEffect(() => {
    fetchStatus();
    const int = setInterval(fetchStatus, 5000);
    return () => clearInterval(int);
  }, []);

<<<<<<< HEAD
  if (loading && !status) return <div className="p-8 text-[var(--text-secondary)] animate-pulse">Loading...</div>;
  if (error && !status) return (
    <div className="p-8 max-w-6xl mx-auto animate-fade-in">
      <div className="surface p-8 border-l-4 border-l-[var(--status-danger)] bg-[var(--status-danger-bg)]">
        <h2 className="text-xl font-bold text-[var(--status-danger)] mb-2">Backend Offline</h2>
        <p className="text-[var(--text-primary)]">The labctl backend could not be reached. Ensure the API server is running.</p>
        <p className="text-sm text-[var(--text-secondary)] mt-2">Error: {error}</p>
      </div>
    </div>
  );

  const cluster = status?.cluster;
=======
  if (loading || !status) return <div className="p-8 text-[var(--text-secondary)]">Loading...</div>;

  const cluster = status.cluster;
>>>>>>> 9b97903 (Fixing conflict)
  
  return (
    <div className="p-8 max-w-6xl mx-auto flex flex-col gap-8 animate-fade-in">
      <div className="flex items-center justify-between">
<<<<<<< HEAD
        <div className="flex items-center gap-4">
          <h2 className="text-2xl font-bold">Cluster Overview</h2>
          {error && <span className="badge badge-danger">Offline</span>}
        </div>
        <div className="badge badge-info">
          {status?.domainSuffix || 'local'}
=======
        <h2 className="text-2xl font-bold">Cluster Overview</h2>
        <div className="badge badge-info">
          {status.domainSuffix}
>>>>>>> 9b97903 (Fixing conflict)
        </div>
      </div>

      {/* Cluster Info Grid */}
      <div className="grid grid-cols-3 gap-6">
<<<<<<< HEAD
        <StatCard icon={<Server />} label="Runtime" value={cluster?.provider || 'Unknown'} />
        <StatCard icon={<Activity />} label="K8s Version" value={cluster?.version || 'Unknown'} />
        <StatCard icon={<Cpu />} label="Nodes" value={cluster?.nodes?.toString() || '0'} />
=======
        <StatCard icon={<Server />} label="Runtime" value={cluster.provider} />
        <StatCard icon={<Activity />} label="K8s Version" value={cluster.version} />
        <StatCard icon={<Cpu />} label="Nodes" value={cluster.nodes.toString()} />
>>>>>>> 9b97903 (Fixing conflict)
      </div>

      {/* Platform Status */}
      <div>
        <h3 className="text-xl font-semibold mb-4 text-[var(--text-primary)]">Platform Health</h3>
        <div className="grid grid-cols-4 gap-4">
<<<<<<< HEAD
          <ComponentCard title="Ingress" status={status?.platform.ingress.active || false} provider={status?.platform.ingress.provider || ''} />
          <ComponentCard title="Metrics" status={status?.platform.metrics.active || false} provider={status?.platform.metrics.provider || ''} />
          <ComponentCard title="Logging" status={status?.platform.logging.active || false} provider={status?.platform.logging.provider || ''} />
          <ComponentCard title="Tracing" status={status?.platform.tracing.active || false} provider={status?.platform.tracing.provider || ''} />
=======
          <ComponentCard title="Ingress" status={status.platform.ingress.active} provider={status.platform.ingress.provider} />
          <ComponentCard title="Metrics" status={status.platform.metrics.active} provider={status.platform.metrics.provider} />
          <ComponentCard title="Logging" status={status.platform.logging.active} provider={status.platform.logging.provider} />
          <ComponentCard title="Tracing" status={status.platform.tracing.active} provider={status.platform.tracing.provider} />
>>>>>>> 9b97903 (Fixing conflict)
        </div>
      </div>

      {/* Apps Summary */}
      <div>
         <h3 className="text-xl font-semibold mb-4 text-[var(--text-primary)]">Applications</h3>
         <div className="surface p-4 flex flex-col gap-3">
<<<<<<< HEAD
           {(!status?.apps || status.apps.length === 0) && <div className="text-[var(--text-muted)] text-sm">No apps registered.</div>}
           {status?.apps?.map(app => (
             <div key={app.name} className="flex items-center justify-between p-3 rounded-[var(--radius-sm)] bg-[var(--bg-base)] border border-[var(--border-light)] hover:border-[var(--border-highlight)] transition-colors">
=======
           {status.apps.length === 0 && <div className="text-[var(--text-muted)] text-sm">No apps registered.</div>}
           {status.apps.map(app => (
             <div key={app.name} className="flex items-center justify-between p-3 rounded-[var(--radius-sm)] bg-[var(--bg-base)] border border-[var(--border-light)]">
>>>>>>> 9b97903 (Fixing conflict)
               <div className="flex items-center gap-3">
                 <Package className="text-[var(--text-muted)]" size={18} />
                 <span className="font-medium">{app.name}</span>
               </div>
               <div className="flex items-center gap-4 text-sm">
                 <span className="text-[var(--text-secondary)]">Build: {app.buildStrategy}</span>
                 <span className="text-[var(--text-secondary)]">Deploy: {app.deployStrategy}</span>
                 {app.deployed ? (
                   <span className="badge badge-success">Running ({app.ready}/{app.replicas})</span>
                 ) : (
                   <span className="badge badge-warning">Not Deployed</span>
                 )}
               </div>
             </div>
           ))}
         </div>
      </div>
    </div>
  );
}

function StatCard({ icon, label, value }: { icon: React.ReactNode, label: string, value: string }) {
  return (
<<<<<<< HEAD
    <div className="surface surface-interactive p-6 flex items-center gap-4">
=======
    <div className="surface p-6 flex items-center gap-4">
>>>>>>> 9b97903 (Fixing conflict)
      <div className="p-3 rounded-full bg-[var(--bg-base)] text-[var(--accent-primary)]">
        {icon}
      </div>
      <div>
        <div className="text-sm text-[var(--text-secondary)] font-medium mb-1">{label}</div>
        <div className="text-xl font-bold text-[var(--text-primary)]">{value}</div>
      </div>
    </div>
  );
}

function ComponentCard({ title, status, provider }: { title: string, status: boolean, provider: string }) {
  return (
<<<<<<< HEAD
    <div className={`surface surface-interactive p-4 border-t-2 ${status ? 'border-t-[var(--status-success)]' : 'border-t-[var(--status-danger)]'}`}>
=======
    <div className={`surface p-4 border-t-2 ${status ? 'border-t-[var(--status-success)]' : 'border-t-[var(--status-danger)]'}`}>
>>>>>>> 9b97903 (Fixing conflict)
      <div className="font-semibold text-lg mb-1">{title}</div>
      <div className="text-sm text-[var(--text-secondary)] mb-3">{provider || 'None'}</div>
      {status ? (
        <span className="badge badge-success text-xs">Healthy</span>
      ) : (
        <span className="badge badge-danger text-xs">Offline</span>
      )}
    </div>
  );
}
