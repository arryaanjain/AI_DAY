import React from 'react';
import { Link } from 'react-router-dom';
import type { PublicConfig, CurrentUser } from '../api/types';
import { Shell } from '../components/layout/Shell';
import { HeroCard, StatTile, Card } from '../components/ui/Card';
import { Button } from '../components/ui/Button';
import { formatMoney } from '../api/client';
import { Sparkles, BookOpen, Layers, ArrowRight, ShieldCheck, Zap, Coins } from 'lucide-react';

interface LandingPageProps {
  config: PublicConfig;
  user: CurrentUser | null;
  onLogout: () => void;
}

export const LandingPage: React.FC<LandingPageProps> = ({ config, user, onLogout }) => {
  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      {/* Hero Section */}
      <section className="grid gap-8 py-10 lg:grid-cols-[1.15fr_0.85fr] items-stretch">
        <div className="rounded-[2.5rem] border border-white/10 bg-slate-900/60 p-8 sm:p-12 backdrop-blur-2xl flex flex-col justify-between shadow-2xl shadow-cyan-950/20">
          <div>
            <div className="inline-flex items-center gap-2 rounded-full border border-cyan-300/30 bg-cyan-300/10 px-4 py-1.5 text-xs font-bold uppercase tracking-widest text-cyan-300 mb-6">
              <Sparkles className="h-3.5 w-3.5" />
              <span>Generative AI Character Engine</span>
            </div>

            <h1 className="text-4xl sm:text-6xl font-black tracking-tight text-white leading-[1.1]">
              Turn your selfies into{' '}
              <span className="bg-gradient-to-r from-cyan-300 via-indigo-200 to-purple-300 bg-clip-text text-transparent">
                3D Art & Life Comics.
              </span>
            </h1>

            <p className="mt-6 text-base sm:text-lg leading-relaxed text-slate-300 max-w-2xl">
              An end-to-end multi-agent AI pipeline. Upload a selfie photo, describe your life’s highest peaks and lowest struggles, and let autonomous backend LLM agents plan, direct, and compose your personalized comic storybook.
            </p>
          </div>

          <div className="mt-10 pt-8 border-t border-white/10 space-y-6">
            <div className="flex flex-wrap gap-3 text-xs font-semibold text-slate-300">
              <span className="rounded-full border border-white/10 bg-white/5 px-4 py-2 flex items-center gap-1.5">
                <Coins className="h-3.5 w-3.5 text-amber-400" />
                Price: {formatMoney(config.pricePerGenerationPaise, config.currency)} / generation
              </span>
              <span className="rounded-full border border-white/10 bg-white/5 px-4 py-2 flex items-center gap-1.5">
                <Zap className="h-3.5 w-3.5 text-cyan-400" />
                Modules: {config.enabledModules.join(', ')}
              </span>
              <span className="rounded-full border border-white/10 bg-white/5 px-4 py-2 flex items-center gap-1.5">
                <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
                Auth Mode: {config.authMode === 'phone' ? 'SMS OTP' : 'Microsoft Email'}
              </span>
            </div>

            <div className="flex flex-wrap gap-4">
              {user ? (
                <Link to="/comic">
                  <Button size="lg" leftIcon={<BookOpen className="h-5 w-5" />} rightIcon={<ArrowRight className="h-4 w-4" />}>
                    Create My Comic Storybook
                  </Button>
                </Link>
              ) : (
                <Link to="/login">
                  <Button size="lg" leftIcon={<Sparkles className="h-5 w-5" />} rightIcon={<ArrowRight className="h-4 w-4" />}>
                    Get Started & Sign In
                  </Button>
                </Link>
              )}

              <Link to="/generations">
                <Button variant="secondary" size="lg" leftIcon={<Layers className="h-5 w-5" />}>
                  View My Creations
                </Button>
              </Link>
            </div>
          </div>
        </div>

        {/* Feature Cards Grid */}
        <div className="grid gap-5">
          <Card className="p-6">
            <p className="text-xs uppercase tracking-widest font-bold text-slate-400">Current Session</p>
            <h2 className="mt-2 text-2xl font-bold text-white">
              {user ? user.displayName : 'Guest User'}
            </h2>
            <p className="mt-1 text-xs text-slate-300">
              {user
                ? (user.phone || user.email || user.id)
                : 'Sign in to reserve generation credits and execute worker jobs.'}
            </p>
          </Card>

          <HeroCard
            title="PixArt 3D Generator"
            description="Upload your front-facing selfie photo to transform it into a stylized 3D avatar portrait."
            action="Launch PixArt"
            href="/pixart"
            icon={<Sparkles className="h-6 w-6" />}
            badge="Fast Pipeline"
          />

          <HeroCard
            title="Personalized Storybook"
            description="Submit your life’s biggest high and low moments. AI agents draft narrative script, panel layout, and print-ready PDF."
            action="Launch Comic Generator"
            href="/comic"
            icon={<BookOpen className="h-6 w-6" />}
            badge="Multi-Agent Pipeline"
          />
        </div>
      </section>

      {/* Stats Summary */}
      <section className="grid gap-5 md:grid-cols-3 pt-6">
        <StatTile
          label="Authentication"
          value={config.authMode === 'phone' ? 'SMS OTP' : 'Email SSO'}
          hint="Configured dynamically via backend AUTH_MODE env"
          icon={<ShieldCheck className="h-5 w-5" />}
        />
        <StatTile
          label="AI Generation Rate"
          value={formatMoney(config.pricePerGenerationPaise, config.currency)}
          hint="1 Credit reserved per successful job submission"
          icon={<Coins className="h-5 w-5" />}
        />
        <StatTile
          label="Active Modules"
          value={`${config.enabledModules.length} Enabled`}
          hint="PixArt 3D Portrait & Multi-Page Comic Storybook"
          icon={<Zap className="h-5 w-5" />}
        />
      </section>
    </Shell>
  );
};
