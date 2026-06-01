import React, { useEffect } from 'react';
import { X } from 'lucide-react';

interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title: string;
  children: React.ReactNode;
}

export function Modal({ isOpen, onClose, title, children }: ModalProps) {
  useEffect(() => {
    if (isOpen) document.body.style.overflow = 'hidden';
    else document.body.style.overflow = 'unset';
    return () => { document.body.style.overflow = 'unset'; }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center animate-fade-in">
      <div 
        className="absolute inset-0 bg-[#050505] bg-opacity-70 backdrop-blur-sm transition-opacity"
        onClick={onClose}
      ></div>
      <div className="relative z-10 w-full max-w-md surface-glass shadow-[0_20px_60px_rgba(0,0,0,0.8)] border border-[var(--border-light)] flex flex-col scale-100 transition-transform">
        <div className="flex items-center justify-between p-5 border-b border-[var(--border-light)] bg-[var(--bg-surface-elevated)] rounded-t-[var(--radius-lg)]">
          <h3 className="text-lg font-bold text-[var(--text-primary)]">{title}</h3>
          <button 
            onClick={onClose}
            className="text-[var(--text-muted)] hover:text-[var(--text-primary)] transition-colors p-1"
          >
            <X size={20} />
          </button>
        </div>
        <div className="p-6">
          {children}
        </div>
      </div>
    </div>
  );
}
