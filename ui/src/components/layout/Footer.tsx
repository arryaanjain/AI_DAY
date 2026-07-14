import React from 'react';
import { Sparkles, Heart } from 'lucide-react';

export const Footer: React.FC = () => {
  return (
    <footer className="mt-auto border-t border-white/10 bg-slate-950/60 py-8 backdrop-blur-md">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs text-slate-400">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-cyan-300" />
          <span className="font-semibold text-slate-200">AI Day</span>
          <span>— Personal AI Character & Narrative Platform</span>
        </div>
        <div className="flex items-center gap-1">
          <span>Crafted with</span>
          <Heart className="h-3.5 w-3.5 text-rose-400 fill-rose-400" />
          <span>for generative storytelling</span>
        </div>
      </div>
    </footer>
  );
};
