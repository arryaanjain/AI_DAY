import React from 'react';
import type { JobStatus } from '../../api/types';
import { Loader2, CheckCircle2, AlertCircle, Clock } from 'lucide-react';

interface BadgeProps {
  status: JobStatus | string;
  size?: 'sm' | 'md';
}

export const Badge: React.FC<BadgeProps> = ({ status, size = 'md' }) => {
  const normalized = status.toLowerCase();

  let style = 'bg-slate-800 text-slate-300 border-slate-700';
  let icon = <Clock className="h-3 w-3" />;
  let label = status;

  if (normalized === 'queued' || normalized === 'created') {
    style = 'bg-amber-500/10 text-amber-300 border-amber-500/30';
    icon = <Clock className="h-3 w-3" />;
    label = 'Queued';
  } else if (normalized === 'processing') {
    style = 'bg-cyan-500/10 text-cyan-300 border-cyan-500/30';
    icon = <Loader2 className="h-3 w-3 animate-spin" />;
    label = 'Processing';
  } else if (normalized === 'completed') {
    style = 'bg-emerald-500/10 text-emerald-300 border-emerald-500/30';
    icon = <CheckCircle2 className="h-3 w-3" />;
    label = 'Completed';
  } else if (normalized === 'failed') {
    style = 'bg-rose-500/10 text-rose-300 border-rose-500/30';
    icon = <AlertCircle className="h-3 w-3" />;
    label = 'Failed';
  }

  const sizeClasses = size === 'sm' ? 'px-2.5 py-0.5 text-[11px] gap-1' : 'px-3 py-1 text-xs gap-1.5';

  return (
    <span
      className={`inline-flex items-center font-medium rounded-full border backdrop-blur-md uppercase tracking-wider ${style} ${sizeClasses}`}
    >
      {icon}
      <span>{label}</span>
    </span>
  );
};
