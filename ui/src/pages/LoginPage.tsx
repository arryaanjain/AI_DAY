import React, { useEffect } from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import type { PublicConfig, CurrentUser } from '../api/types';
import { Shell } from '../components/layout/Shell';
import { PhoneAuthForm } from '../components/auth/PhoneAuthForm';
import { EmailAuthForm } from '../components/auth/EmailAuthForm';
import { ShieldCheck, Lock } from 'lucide-react';

interface LoginPageProps {
  config: PublicConfig;
  user: CurrentUser | null;
  onLogout: () => void;
}

export const LoginPage: React.FC<LoginPageProps> = ({ config, user, onLogout }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const searchParams = new URLSearchParams(location.search);
  const nextPath = searchParams.get('next') || '/';

  useEffect(() => {
    if (user) {
      navigate(nextPath, { replace: true });
    }
  }, [user, navigate, nextPath]);

  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      <section className="mx-auto max-w-lg py-12">
        <div className="rounded-[2.5rem] border border-white/10 bg-slate-900/70 p-8 sm:p-10 backdrop-blur-2xl shadow-2xl shadow-cyan-950/30">
          <div className="flex items-center gap-3 mb-6">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-cyan-300/10 text-cyan-300 border border-cyan-300/20">
              <Lock className="h-6 w-6" />
            </div>
            <div>
              <span className="text-xs font-bold uppercase tracking-widest text-cyan-300">
                Secure Authentication
              </span>
              <h1 className="text-2xl font-bold text-white">Sign in to AI Day</h1>
            </div>
          </div>

          <p className="text-sm text-slate-300 mb-8 leading-relaxed">
            Authentication mode is controlled dynamically by backend configuration ({config.authMode === 'phone' ? 'Phone SMS OTP' : 'Microsoft Email'}).
          </p>

          {config.authMode === 'phone' ? (
            <PhoneAuthForm onSuccess={() => navigate(nextPath, { replace: true })} />
          ) : (
            <EmailAuthForm nextPath={nextPath} />
          )}

          <div className="mt-8 pt-6 border-t border-white/10 flex items-center justify-center gap-2 text-xs text-slate-400">
            <ShieldCheck className="h-4 w-4 text-emerald-400" />
            <span>Encrypted Session & Credentials</span>
          </div>
        </div>
      </section>
    </Shell>
  );
};
