import React from 'react';
import { Navbar } from './Navbar';
import { Footer } from './Footer';
import type { CurrentUser, AuthMode } from '../../api/types';

interface ShellProps {
  children: React.ReactNode;
  user: CurrentUser | null;
  authMode: AuthMode;
  onLogout: () => void;
}

export const Shell: React.FC<ShellProps> = ({ children, user, authMode, onLogout }) => {
  return (
    <div className="min-h-screen flex flex-col bg-[radial-gradient(circle_at_top,_rgba(34,211,238,0.15),_transparent_40%),radial-gradient(circle_at_bottom_right,_rgba(99,102,241,0.12),_transparent_45%),linear-gradient(180deg,#020617_0%,#0f172a_50%,#020617_100%)] text-slate-100 font-sans selection:bg-cyan-300 selection:text-slate-950">
      <Navbar user={user} authMode={authMode} onLogout={onLogout} />
      <main className="flex-1 mx-auto max-w-7xl w-full px-4 sm:px-6 pb-16 pt-2">
        {children}
      </main>
      <Footer />
    </div>
  );
};
