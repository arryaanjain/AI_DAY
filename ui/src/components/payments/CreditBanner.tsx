import React from 'react';
import { Coins, Plus } from 'lucide-react';
import { Button } from '../ui/Button';

interface CreditBannerProps {
  balance: number;
  onBuyCredits: () => void;
}

export const CreditBanner: React.FC<CreditBannerProps> = ({ balance, onBuyCredits }) => {
  return (
    <div className="rounded-2xl border border-amber-500/30 bg-gradient-to-r from-amber-500/10 via-amber-500/5 to-transparent p-4 backdrop-blur-md flex flex-col sm:flex-row items-center justify-between gap-4">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-2xl bg-amber-500/20 text-amber-300 border border-amber-500/30">
          <Coins className="h-5 w-5" />
        </div>
        <div>
          <h4 className="text-sm font-bold text-white flex items-center gap-2">
            <span>Generation Credits: {balance} Available</span>
            {balance === 0 && (
              <span className="rounded-full bg-rose-500/20 px-2 py-0.5 text-[10px] font-bold text-rose-300 border border-rose-500/30 uppercase">
                Low Balance
              </span>
            )}
          </h4>
          <p className="text-xs text-slate-300">
            {balance > 0
              ? 'You are ready to queue 3D PixArt portraits or storybooks.'
              : 'Add generation credits to unlock instant character & comic generation.'}
          </p>
        </div>
      </div>

      <Button
        variant="primary"
        size="sm"
        onClick={onBuyCredits}
        leftIcon={<Plus className="h-4 w-4" />}
      >
        Buy Credits
      </Button>
    </div>
  );
};
