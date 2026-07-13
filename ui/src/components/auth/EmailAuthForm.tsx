import React, { useState } from 'react';
import { api } from '../../api/client';
import { Input } from '../ui/Input';
import { Button } from '../ui/Button';
import { ArrowRight } from 'lucide-react';

interface EmailAuthFormProps {
  nextPath?: string;
}

export const EmailAuthForm: React.FC<EmailAuthFormProps> = ({ nextPath = '/' }) => {
  const [email, setEmail] = useState('');

  const handleContinue = (e: React.FormEvent) => {
    e.preventDefault();
    const search = new URLSearchParams({ next: nextPath });
    if (email.trim()) search.set('email', email.trim());
    window.location.href = `${api.defaults.baseURL}/api/v1/auth/microsoft/start?${search.toString()}`;
  };

  return (
    <form onSubmit={handleContinue} className="space-y-5">
      <Input
        label="Email Address"
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="user@example.com"
        hint="You will be redirected to complete Microsoft Email authentication."
      />

      <Button
        type="submit"
        size="lg"
        className="w-full"
        rightIcon={<ArrowRight className="h-4 w-4" />}
      >
        Continue with Email
      </Button>

      <p className="text-xs text-slate-400 text-center leading-relaxed">
        This triggers the backend Microsoft SSO authentication flow, which creates an authenticated session cookie upon completion.
      </p>
    </form>
  );
};
