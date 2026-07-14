import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchCreditBalance } from '../../api/client';
import type { CurrentUser, AuthMode } from '../../api/types';
import { Sparkles, Coins, LogOut, LogIn, CreditCard, Layers, BookOpen } from 'lucide-react';

interface NavbarProps {
  user: CurrentUser | null;
  authMode: AuthMode;
  onLogout: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ user, authMode, onLogout }) => {
  const location = useLocation();

  // Fetch credit balance if user is logged in
  const { data: credits = 0 } = useQuery({
    queryKey: ['credit-balance'],
    queryFn: fetchCreditBalance,
    enabled: !!user,
    refetchInterval: 10000,
  });

  const isActive = (path: string) => location.pathname === path;

  const navLinkClass = (path: string) =>
    `inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-sm font-medium transition-colors ${
      isActive(path)
        ? 'bg-white/10 text-white border border-white/10'
        : 'text-slate-300 hover:text-white hover:bg-white/5'
    }`;

  return (
    <header className="mx-auto max-w-7xl px-4 sm:px-6 py-5">
      <nav className="flex items-center justify-between rounded-3xl border border-white/10 bg-slate-900/60 px-5 py-3.5 backdrop-blur-xl shadow-xl">
        <Link to="/" className="flex items-center gap-2.5 text-lg font-bold tracking-tight text-white hover:opacity-90 transition-opacity">
          <div className="flex h-9 w-9 items-center justify-center rounded-2xl bg-gradient-to-tr from-cyan-400 to-indigo-500 text-slate-950 font-black shadow-lg shadow-cyan-500/20">
            <Sparkles className="h-5 w-5 fill-slate-950" />
          </div>
          <span className="bg-gradient-to-r from-white via-slate-100 to-cyan-200 bg-clip-text text-transparent">
            AI Day
          </span>
        </Link>

        <div className="hidden md:flex items-center gap-1">
          <Link to="/pixart" className={navLinkClass('/pixart')}>
            <Sparkles className="h-4 w-4 text-cyan-300" />
            <span>PixArt</span>
          </Link>
          <Link to="/comic" className={navLinkClass('/comic')}>
            <BookOpen className="h-4 w-4 text-indigo-300" />
            <span>Comic</span>
          </Link>
          <Link to="/generations" className={navLinkClass('/generations')}>
            <Layers className="h-4 w-4 text-purple-300" />
            <span>My Creations</span>
          </Link>
          <Link to="/payments" className={navLinkClass('/payments')}>
            <CreditCard className="h-4 w-4 text-emerald-300" />
            <span>Payments</span>
          </Link>
        </div>

        <div className="flex items-center gap-3">
          {user && (
            <Link
              to="/payments"
              className="flex items-center gap-1.5 rounded-full border border-amber-500/30 bg-amber-500/10 px-3 py-1 text-xs font-semibold text-amber-300 hover:bg-amber-500/20 transition-colors"
              title="Available generation credits"
            >
              <Coins className="h-3.5 w-3.5" />
              <span>{credits} {credits === 1 ? 'Credit' : 'Credits'}</span>
            </Link>
          )}

          <span className="hidden sm:inline-flex rounded-full border border-white/10 bg-white/5 px-3 py-1 text-[10px] uppercase font-bold tracking-widest text-slate-300">
            {authMode === 'phone' ? 'SMS Auth' : 'Email Auth'}
          </span>

          {user ? (
            <div className="flex items-center gap-2">
              <span className="hidden lg:inline-block text-xs font-medium text-slate-300 truncate max-w-[120px]">
                {user.displayName || user.phone || user.email}
              </span>
              <button
                onClick={onLogout}
                className="inline-flex items-center gap-1.5 rounded-full bg-cyan-300 hover:bg-cyan-200 px-4 py-1.5 text-xs font-bold text-slate-950 transition-all shadow-md shadow-cyan-500/20 active:scale-95 cursor-pointer"
              >
                <LogOut className="h-3.5 w-3.5" />
                <span>Logout</span>
              </button>
            </div>
          ) : (
            <Link
              to="/login"
              className="inline-flex items-center gap-1.5 rounded-full bg-cyan-300 hover:bg-cyan-200 px-4 py-1.5 text-xs font-bold text-slate-950 transition-all shadow-md shadow-cyan-500/20 active:scale-95"
            >
              <LogIn className="h-3.5 w-3.5" />
              <span>Login</span>
            </Link>
          )}
        </div>
      </nav>

      {/* Mobile nav subbar */}
      <div className="md:hidden mt-3 flex items-center justify-center gap-2 overflow-x-auto py-2">
        <Link to="/pixart" className={navLinkClass('/pixart')}>
          <span>PixArt</span>
        </Link>
        <Link to="/comic" className={navLinkClass('/comic')}>
          <span>Comic</span>
        </Link>
        <Link to="/generations" className={navLinkClass('/generations')}>
          <span>My Creations</span>
        </Link>
        <Link to="/payments" className={navLinkClass('/payments')}>
          <span>Payments</span>
        </Link>
      </div>
    </header>
  );
};
