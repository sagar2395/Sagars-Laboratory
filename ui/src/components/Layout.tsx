import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import { LayoutDashboard, PlayCircle, Layers, Box, Terminal } from 'lucide-react';
import { useActionEvents } from '../api/useActionEvents';

export function Layout() {
  const { isConnected } = useActionEvents(); // keep WS connected globally

  return (
    <div className="flex h-screen w-screen overflow-hidden relative text-[var(--text-primary)]">
      {/* Background Mesh Gradient */}
      <div className="mesh-bg"></div>

      {/* Floating Sidebar */}
      <aside className="w-64 flex-shrink-0 flex flex-col surface-glass m-6 z-10 border border-[var(--border-light)] shadow-[var(--shadow-glass)]">
        <div className="p-6 flex items-center gap-3 border-b border-[var(--border-light)] bg-[var(--bg-surface)] rounded-t-[var(--radius-lg)]">
          <Terminal size={24} className="text-[var(--accent-primary)] drop-shadow-[0_0_8px_var(--accent-glow)]" />
          <h1 className="text-xl font-bold tracking-tight">labctl</h1>
          <div title={isConnected ? "WebSocket Connected" : "WebSocket Disconnected"} 
               className={`ml-auto w-2.5 h-2.5 rounded-full ${isConnected ? 'status-dot-success' : 'bg-[var(--status-danger)]'}`}>
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
      <main className="flex-1 overflow-auto relative z-10 py-6 pr-6 animate-fade-in">
        <div className="h-full w-full bg-[var(--bg-surface)] backdrop-blur-md rounded-[var(--radius-lg)] border border-[var(--border-light)] shadow-[var(--shadow-md)] overflow-auto">
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
        flex items-center gap-3 p-3 rounded-[var(--radius-md)] transition-all duration-300
        ${isActive 
          ? 'bg-[var(--accent-primary)] bg-opacity-20 text-[var(--accent-primary)] border border-[var(--accent-glow)] shadow-[0_0_15px_var(--accent-glow)]' 
          : 'text-[var(--text-secondary)] border border-transparent hover:bg-[var(--bg-surface-elevated)] hover:text-[var(--text-primary)] hover:border-[var(--border-highlight)]'}
      `}
    >
      {icon}
      <span className="font-medium tracking-wide">{label}</span>
    </NavLink>
  );
}
