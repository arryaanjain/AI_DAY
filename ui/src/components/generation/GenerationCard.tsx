import React from 'react';
import type { GenerationJob } from '../../api/types';
import { Badge } from '../ui/Badge';
import { Button } from '../ui/Button';
import { Sparkles, BookOpen, Clock, ArrowRight } from 'lucide-react';
import { Link } from 'react-router-dom';

interface GenerationCardProps {
  job: GenerationJob;
  onSelect?: (job: GenerationJob) => void;
}

export const GenerationCard: React.FC<GenerationCardProps> = ({ job, onSelect }) => {
  const isComic = job.module === 'comic';

  const dateStr = new Date(job.createdAt).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div className="rounded-[1.75rem] border border-white/10 bg-slate-900/60 p-5 backdrop-blur-xl transition-all duration-200 hover:border-cyan-300/30 hover:bg-slate-900/80 flex flex-col justify-between gap-4">
      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <div className="flex items-center gap-2">
            {isComic ? (
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                <BookOpen className="h-4 w-4" />
              </div>
            ) : (
              <div className="flex h-8 w-8 items-center justify-center rounded-xl bg-cyan-500/10 text-cyan-300 border border-cyan-500/20">
                <Sparkles className="h-4 w-4" />
              </div>
            )}
            <div>
              <h4 className="text-sm font-bold text-white">
                {isComic ? 'Personal Comic Storybook' : 'PixArt 3D Portrait'}
              </h4>
              <p className="text-[11px] text-slate-400 flex items-center gap-1">
                <Clock className="h-3 w-3" />
                <span>{dateStr}</span>
              </p>
            </div>
          </div>
          <Badge status={job.status} size="sm" />
        </div>

        {job.input && isComic && (
          <div className="text-xs text-slate-300 rounded-xl bg-slate-950/60 p-3 border border-white/5 space-y-1">
            <p className="line-clamp-1 font-medium">
              <span className="text-slate-400">High:</span> {job.input.biggestHigh}
            </p>
            <p className="line-clamp-1 font-medium">
              <span className="text-slate-400">Low:</span> {job.input.biggestLow}
            </p>
          </div>
        )}
      </div>

      <div className="flex items-center justify-between pt-2 border-t border-white/5">
        <span className="text-[11px] font-mono text-slate-500">
          ID: {job.id.substring(0, 8)}…
        </span>

        {onSelect ? (
          <Button variant="ghost" size="sm" onClick={() => onSelect(job)} rightIcon={<ArrowRight className="h-3.5 w-3.5" />}>
            View Details
          </Button>
        ) : (
          <Link to={`/generations/${job.id}`}>
            <Button variant="ghost" size="sm" rightIcon={<ArrowRight className="h-3.5 w-3.5" />}>
              View Details
            </Button>
          </Link>
        )}
      </div>
    </div>
  );
};
