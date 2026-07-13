import React, { useState, useEffect } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { sendPhoneOTP, resendPhoneOTP, verifyPhoneOTP, getApiErrorMessage } from '../../api/client';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { ArrowRight, CheckCircle2, RefreshCw } from 'lucide-react';

interface PhoneAuthFormProps {
  onSuccess: () => void;
}

export const PhoneAuthForm: React.FC<PhoneAuthFormProps> = ({ onSuccess }) => {
  const queryClient = useQueryClient();
  const [phone, setPhone] = useState('+919876543210');
  const [code, setCode] = useState('');
  const [isSent, setIsSent] = useState(false);
  const [feedback, setFeedback] = useState<{ type: 'success' | 'error'; message: string } | null>(null);
  const [cooldown, setCooldown] = useState(0);

  useEffect(() => {
    let timer: ReturnType<typeof setInterval>;
    if (cooldown > 0) {
      timer = setInterval(() => setCooldown((prev) => prev - 1), 1000);
    }
    return () => clearInterval(timer);
  }, [cooldown]);

  const getFormattedPhone = () => {
    const raw = phone.trim();
    if (!raw) return '';
    if (raw.startsWith('+')) return raw;
    return `+91${raw.replace(/\D/g, '')}`;
  };

  const sendOtpMutation = useMutation({
    mutationFn: () => sendPhoneOTP(getFormattedPhone()),
    onSuccess: (data) => {
      setIsSent(true);
      setCooldown(30);
      setFeedback({
        type: 'success',
        message: data.status === 'otp_sent' ? 'OTP sent successfully to your phone!' : 'OTP code requested.',
      });
    },
    onError: (error) => {
      setFeedback({
        type: 'error',
        message: getApiErrorMessage(error),
      });
    },
  });

  const resendOtpMutation = useMutation({
    mutationFn: () => resendPhoneOTP(getFormattedPhone()),
    onSuccess: () => {
      setCooldown(30);
      setFeedback({
        type: 'success',
        message: 'A new OTP code has been sent.',
      });
    },
    onError: (error) => {
      setFeedback({
        type: 'error',
        message: getApiErrorMessage(error),
      });
    },
  });

  const verifyOtpMutation = useMutation({
    mutationFn: () => verifyPhoneOTP(getFormattedPhone(), code.trim()),
    onSuccess: async () => {
      await queryClient.refetchQueries({ queryKey: ['me'] });
      await queryClient.refetchQueries({ queryKey: ['credit-balance'] });
      onSuccess();
    },
    onError: (error) => {
      setFeedback({
        type: 'error',
        message: getApiErrorMessage(error),
      });
    },
  });

  const handleSendOrVerify = async (e: React.FormEvent) => {
    e.preventDefault();
    setFeedback(null);
    if (!isSent) {
      if (!phone.trim()) {
        setFeedback({ type: 'error', message: 'Please enter a valid mobile phone number.' });
        return;
      }
      sendOtpMutation.mutate();
    } else {
      if (!code.trim()) {
        setFeedback({ type: 'error', message: 'Please enter the 6-digit OTP code.' });
        return;
      }
      verifyOtpMutation.mutate();
    }
  };

  return (
    <form onSubmit={handleSendOrVerify} className="space-y-5">
      <Input
        label="Mobile Phone Number"
        type="tel"
        value={phone}
        onChange={(e) => setPhone(e.target.value)}
        placeholder="+919876543210"
        disabled={isSent || sendOtpMutation.isPending || verifyOtpMutation.isPending}
        hint="Include country code (e.g. +91 for India)"
      />

      {isSent && (
        <div className="animate-in fade-in slide-in-from-top-2 duration-300">
          <Input
            label="Verification Code (OTP)"
            type="text"
            maxLength={6}
            value={code}
            onChange={(e) => setCode(e.target.value)}
            placeholder="123456"
            disabled={verifyOtpMutation.isPending}
            hint="Check backend terminal logs in development mode for OTP code"
          />
        </div>
      )}

      {feedback && (
        <div
          className={`flex items-center gap-2 rounded-2xl p-4 text-xs font-medium border ${
            feedback.type === 'success'
              ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
              : 'border-rose-500/30 bg-rose-500/10 text-rose-300'
          }`}
        >
          <CheckCircle2 className="h-4 w-4 shrink-0" />
          <span>{feedback.message}</span>
        </div>
      )}

      <div className="pt-2">
        {!isSent ? (
          <Button
            type="submit"
            size="lg"
            className="w-full"
            isLoading={sendOtpMutation.isPending}
            rightIcon={<ArrowRight className="h-4 w-4" />}
          >
            Send OTP Code
          </Button>
        ) : (
          <div className="space-y-3">
            <Button
              type="submit"
              size="lg"
              className="w-full"
              isLoading={verifyOtpMutation.isPending}
              rightIcon={<ArrowRight className="h-4 w-4" />}
            >
              Verify & Sign In
            </Button>

            <div className="flex items-center justify-between text-xs text-slate-400">
              <button
                type="button"
                onClick={() => {
                  setIsSent(false);
                  setCode('');
                  setFeedback(null);
                }}
                className="hover:text-white underline"
              >
                Change Phone Number
              </button>

              <button
                type="button"
                disabled={cooldown > 0 || resendOtpMutation.isPending}
                onClick={() => resendOtpMutation.mutate()}
                className="inline-flex items-center gap-1 hover:text-cyan-300 disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              >
                <RefreshCw className={`h-3 w-3 ${resendOtpMutation.isPending ? 'animate-spin' : ''}`} />
                <span>{cooldown > 0 ? `Resend in ${cooldown}s` : 'Resend OTP'}</span>
              </button>
            </div>
          </div>
        )}
      </div>
    </form>
  );
};
