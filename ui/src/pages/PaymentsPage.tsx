import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { PublicConfig, CurrentUser } from '../api/types';
import { fetchCreditBalance, formatMoney } from '../api/client';
import { Shell } from '../components/layout/Shell';
import { PaymentModal } from '../components/payments/PaymentModal';
import { Button } from '../components/ui/Button';
import { CreditCard, Coins, Plus, ShieldCheck, Zap } from 'lucide-react';

interface PaymentsPageProps {
  config: PublicConfig;
  user: CurrentUser;
  onLogout: () => void;
}

export const PaymentsPage: React.FC<PaymentsPageProps> = ({ config, user, onLogout }) => {
  const [isPaymentModalOpen, setIsPaymentModalOpen] = useState(false);

  const { data: credits = 0, refetch } = useQuery({
    queryKey: ['credit-balance'],
    queryFn: fetchCreditBalance,
    refetchInterval: 5000,
  });

  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      <div className="space-y-8 py-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-black text-white flex items-center gap-3">
              <CreditCard className="h-8 w-8 text-emerald-300" />
              <span>Payments & Credit Balance</span>
            </h1>
            <p className="mt-1 text-sm text-slate-300">
              Manage generation credits, view Razorpay pricing tiers, and top up your account.
            </p>
          </div>

          <Button
            size="lg"
            onClick={() => setIsPaymentModalOpen(true)}
            leftIcon={<Plus className="h-5 w-5" />}
          >
            Buy Generation Credits
          </Button>
        </div>

        {/* Balance Card & Pricing */}
        <div className="grid gap-6 lg:grid-cols-[0.8fr_1.2fr]">
          <div className="rounded-[2.5rem] border border-amber-500/30 bg-gradient-to-br from-amber-500/15 via-slate-900/80 to-slate-950 p-8 backdrop-blur-2xl flex flex-col justify-between shadow-2xl shadow-amber-950/20">
            <div>
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold uppercase tracking-widest text-amber-300">
                  Account Balance
                </span>
                <Coins className="h-6 w-6 text-amber-400 animate-bounce" />
              </div>

              <div className="mt-6">
                <p className="text-5xl font-black text-white">{credits}</p>
                <p className="mt-1 text-sm font-semibold text-amber-300">
                  {credits === 1 ? 'Credit Available' : 'Credits Available'}
                </p>
              </div>
            </div>

            <div className="mt-8 pt-6 border-t border-white/10 space-y-4">
              <p className="text-xs text-slate-300 leading-relaxed">
                1 Credit is automatically reserved per successful generation job submission.
              </p>

              <Button
                variant="outline"
                size="md"
                className="w-full border-amber-400/40 text-amber-300 hover:bg-amber-400/10"
                onClick={() => setIsPaymentModalOpen(true)}
                leftIcon={<Plus className="h-4 w-4" />}
              >
                Top Up Balance
              </Button>
            </div>
          </div>

          {/* Pricing Info */}
          <div className="rounded-[2.5rem] border border-white/10 bg-slate-900/60 p-8 backdrop-blur-2xl space-y-6">
            <div>
              <h3 className="text-xl font-bold text-white">Pricing & Credit Ledger</h3>
              <p className="text-xs text-slate-300 mt-1">
                Standard pricing is configured directly via backend public settings.
              </p>
            </div>

            <div className="space-y-3">
              <div className="flex items-center justify-between p-4 rounded-2xl border border-white/5 bg-slate-950/60">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                    <Zap className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-bold text-white">Standard Generation Rate</p>
                    <p className="text-xs text-slate-400">3D PixArt Portrait or Comic Storybook</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-base font-bold text-cyan-300">
                    {formatMoney(config.pricePerGenerationPaise, config.currency)}
                  </p>
                  <p className="text-[10px] text-slate-400">per generation</p>
                </div>
              </div>

              <div className="flex items-center justify-between p-4 rounded-2xl border border-white/5 bg-slate-950/60">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
                    <ShieldCheck className="h-5 w-5" />
                  </div>
                  <div>
                    <p className="text-sm font-bold text-white">Razorpay Order Gateway</p>
                    <p className="text-xs text-slate-400">Encrypted merchant keys & webhook confirmation</p>
                  </div>
                </div>
                <span className="rounded-full bg-emerald-500/20 px-2.5 py-1 text-[10px] font-bold text-emerald-300 border border-emerald-500/30">
                  Verified
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <PaymentModal
        isOpen={isPaymentModalOpen}
        onClose={() => setIsPaymentModalOpen(false)}
        config={config}
        onPaymentSuccess={refetch}
      />
    </Shell>
  );
};
