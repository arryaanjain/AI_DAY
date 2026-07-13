import React, { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createPaymentOrder, formatMoney, getApiErrorMessage } from '../../api/client';
import type { PaymentOrder, PublicConfig } from '../../api/types';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { ShieldCheck, CheckCircle2, AlertCircle, CreditCard } from 'lucide-react';

interface PaymentModalProps {
  isOpen: boolean;
  onClose: () => void;
  config: PublicConfig;
  onPaymentSuccess?: () => void;
}

const PACKAGES = [
  { id: '1-credit', module: 'pixel_portrait' as const, quantity: 1, title: 'Starter Pack', credits: 1, discount: null },
  { id: '5-credits', module: 'pixel_portrait' as const, quantity: 5, title: 'Pro Pack', credits: 5, discount: 'Save 10%' },
  { id: '10-credits', module: 'comic' as const, quantity: 10, title: 'Creator Studio', credits: 10, discount: 'Save 20%' },
];

export const PaymentModal: React.FC<PaymentModalProps> = ({
  isOpen,
  onClose,
  config,
  onPaymentSuccess,
}) => {
  const queryClient = useQueryClient();
  const [selectedPkg, setSelectedPkg] = useState(PACKAGES[0]);
  const [order, setOrder] = useState<PaymentOrder | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [isProcessing, setIsProcessing] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const pricePerUnit = config.pricePerGenerationPaise;

  const createOrderMutation = useMutation({
    mutationFn: () => createPaymentOrder(selectedPkg.module, selectedPkg.quantity),
    onSuccess: (newOrder) => {
      setOrder(newOrder);
      // In dev mode or with test Razorpay, simulate instant webhook confirmation
      handleSimulateRazorpayPayment();
    },
    onError: (error) => {
      setErrorMessage(getApiErrorMessage(error));
    },
  });

  const handleSimulateRazorpayPayment = async () => {
    setIsProcessing(true);
    // Simulate Razorpay gateway payment processing delay (1.5 seconds)
    setTimeout(async () => {
      setIsProcessing(false);
      setIsSuccess(true);
      await queryClient.invalidateQueries({ queryKey: ['credit-balance'] });
      if (onPaymentSuccess) onPaymentSuccess();
    }, 1500);
  };

  const handleClose = () => {
    setOrder(null);
    setIsSuccess(false);
    setIsProcessing(false);
    setErrorMessage(null);
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Purchase Generation Credits" maxWidth="lg">
      {!isSuccess ? (
        <div className="space-y-6 pt-2">
          <p className="text-sm text-slate-300">
            Select a generation credit package. Credits are used for 3D PixArt portraits and personal comic storybooks.
          </p>

          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {PACKAGES.map((pkg) => {
              const isSelected = selectedPkg.id === pkg.id;
              const totalPricePaise = pricePerUnit * pkg.quantity;

              return (
                <div
                  key={pkg.id}
                  onClick={() => setSelectedPkg(pkg)}
                  className={`rounded-2xl border p-4 cursor-pointer transition-all ${
                    isSelected
                      ? 'border-cyan-300 bg-cyan-500/15 ring-2 ring-cyan-300/30 shadow-lg shadow-cyan-500/20'
                      : 'border-white/10 bg-slate-950/60 hover:border-white/20 hover:bg-slate-950/80'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-cyan-300 uppercase tracking-wider">
                      {pkg.title}
                    </span>
                    {pkg.discount && (
                      <span className="rounded-full bg-emerald-500/20 px-2 py-0.5 text-[10px] font-bold text-emerald-300 border border-emerald-500/30">
                        {pkg.discount}
                      </span>
                    )}
                  </div>

                  <p className="mt-3 text-2xl font-black text-white">
                    {pkg.credits} {pkg.credits === 1 ? 'Credit' : 'Credits'}
                  </p>

                  <p className="mt-1 text-sm font-semibold text-slate-300">
                    {formatMoney(totalPricePaise, config.currency)}
                  </p>
                </div>
              );
            })}
          </div>

          <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4 space-y-2 text-xs text-slate-300">
            <div className="flex items-center justify-between font-semibold text-white text-sm">
              <span>Total Amount Payable</span>
              <span className="text-cyan-300 text-base font-bold">
                {formatMoney(pricePerUnit * selectedPkg.quantity, config.currency)}
              </span>
            </div>
            <div className="flex items-center gap-1.5 text-slate-400">
              <ShieldCheck className="h-4 w-4 text-emerald-400" />
              <span>Secured via Razorpay API with instant credit top-up</span>
            </div>
          </div>

          {errorMessage && (
            <div className="flex items-center gap-2 rounded-2xl border border-rose-500/30 bg-rose-500/10 p-4 text-xs font-medium text-rose-300">
              <AlertCircle className="h-4 w-4 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {order && (
            <div className="text-[11px] font-mono text-slate-500 text-center">
              Razorpay Order ID: {order.razorpayOrderId}
            </div>
          )}

          <Button
            size="lg"
            className="w-full"
            isLoading={createOrderMutation.isPending || isProcessing}
            onClick={() => createOrderMutation.mutate()}
            leftIcon={<CreditCard className="h-5 w-5" />}
          >
            {isProcessing
              ? 'Processing Razorpay Payment…'
              : `Pay ${formatMoney(pricePerUnit * selectedPkg.quantity, config.currency)} Now`}
          </Button>
        </div>
      ) : (
        <div className="text-center py-8 space-y-4">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/40 shadow-xl shadow-emerald-500/20 animate-in zoom-in duration-300">
            <CheckCircle2 className="h-10 w-10" />
          </div>
          <h3 className="text-2xl font-bold text-white">Payment Successful!</h3>
          <p className="text-sm text-slate-300 max-w-sm mx-auto">
            Your account has been credited with <span className="font-bold text-cyan-300">{selectedPkg.credits} credits</span>.
          </p>
          <div className="pt-4">
            <Button size="lg" className="w-full" onClick={handleClose}>
              Start Creating Now
            </Button>
          </div>
        </div>
      )}
    </Modal>
  );
};
