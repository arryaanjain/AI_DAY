import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { PublicConfig, CurrentUser } from '../api/types';
import { fetchCreditBalance } from '../api/client';
import { Shell } from '../components/layout/Shell';
import { PixartForm } from '../components/generation/PixartForm';
import { CreditBanner } from '../components/payments/CreditBanner';
import { PaymentModal } from '../components/payments/PaymentModal';
import { Sparkles, ArrowLeft } from 'lucide-react';
import { Link } from 'react-router-dom';

interface PixartPageProps {
  config: PublicConfig;
  user: CurrentUser;
  onLogout: () => void;
}

export const PixartPage: React.FC<PixartPageProps> = ({ config, user, onLogout }) => {
  const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);

  const { data: credits = 0 } = useQuery({
    queryKey: ['credit-balance'],
    queryFn: fetchCreditBalance,
    refetchInterval: 5000,
  });

  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      <div className="space-y-8 py-6">
        {/* Top Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <Link to="/" className="inline-flex items-center gap-1 text-xs font-semibold text-cyan-300 hover:underline mb-2">
              <ArrowLeft className="h-3.5 w-3.5" />
              <span>Back to Dashboard</span>
            </Link>
            <h1 className="text-3xl font-black text-white flex items-center gap-3">
              <Sparkles className="h-8 w-8 text-cyan-300" />
              <span>PixArt 3D Avatar Generator</span>
            </h1>
            <p className="mt-1 text-sm text-slate-300 max-w-xl">
              Transform your front-facing selfie into a stylized 3D portrait character.
            </p>
          </div>
        </div>

        {/* Credit Banner */}
        <CreditBanner balance={credits} onBuyCredits={() => setIsPaymentModalOpen(true)} />

        {/* Main Form Container */}
        <div className="rounded-[2.5rem] border border-white/10 bg-slate-900/60 p-6 sm:p-10 backdrop-blur-2xl shadow-2xl shadow-cyan-950/20 max-w-3xl">
          <PixartForm />
        </div>
      </div>

      {/* Payment Modal */}
      <PaymentModal
        isOpen={isPaymentModalOpen}
        onClose={() => setIsPaymentModalOpen(false)}
        config={config}
      />
    </Shell>
  );
};
