import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { LayoutDashboard, PlayCircle, Layers, Box, Terminal } from 'lucide-react';
import { useActionEvents } from '../api/useActionEvents';

export function Layout() {
  const { isConnected } = useActionEvents(); // keep WS connected globally

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-[var(--bg-base)]">
      {/* Sidebar */}
      <aside className="w-64 flex-shrink-0 flex flex-col surface-glass m-4 z-10 border-r border-transparent">
        <div className="p-6 flex items-center gap-3 border-b border-[var(--border-light)]">
          <Terminal size={24} className="text-[var(--accent-primary)]" />
          <h1 className="text-xl font-bold tracking-tight">labctl</h1>
          <div title={isConnected ? "WebSocket Connected" : "WebSocket Disconnected"} 
               className={`ml-auto w-2 h-2 rounded-full ${isConnected ? 'bg-[var(--status-success)] shadow-[0_0_8px_var(--status-success)]' : 'bg-[var(--status-danger)]'}`}>
          </div>
        </div>
        
        <nav className="flex-1 p-4 flex flex-col gap-2">
          <NavItem to="/" icon={<LayoutDashboard size={20} />} label="Dashboard" />
          <NavItem to="/scenarios" icon={<PlayCircle size={20} />} label="Scenarios" />
          <NavItem to="/platform" icon={<Layers size={20} />} label="Platform" />
          <NavItem to="/apps" icon={<Box size={20} />} label="Apps" />
        </nav>
      </aside>

      {/* Main Content Area */}
      <main className="flex-1 overflow-auto relative animate-fade-in py-4 pr-4">
        <div className="h-full w-full bg-[var(--bg-surface)] rounded-[var(--radius-lg)] border border-[var(--border-light)] shadow-[var(--shadow-md)] overflow-auto">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

function NavItem({ to, icon, label }: { to: string, icon: React.ReactNode, label: string }) {
  return (
    <NavLink 
      to={to} 
      className={({isActive}) => `
        flex items-center gap-3 p-3 rounded-[var(--radius-md)] transition-all
        ${isActive 
          ? 'bg-[var(--accent-primary)] bg-opacity-20 text-[var(--accent-primary)] shadow-[inset_4px_0_0_var(--accent-primary)]' 
          : 'text-[var(--text-secondary)] hover:bg-[var(--bg-surface-elevated)] hover:text-[var(--text-primary)]'}
      `}
    >
      {icon}
      <span className="font-medium">{label}</span>
    </NavLink>
  );
}
