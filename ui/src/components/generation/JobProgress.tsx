import React from 'react';
import type { GenerationJob, ComicPipelineState } from '../../api/types';
import { Badge } from '../ui/Badge';
import { AlertCircle, Sparkles, BookOpen, FileText, Palette, Image as ImageIcon, FileCheck } from 'lucide-react';

interface JobProgressProps {
  job: GenerationJob;
}

const STAGES = [
  { key: 'narrative_planner', label: '1. Narrative Planner', icon: BookOpen },
  { key: 'safety_review', label: '2. Safety Review', icon: FileText },
  { key: 'art_direction', label: '3. Art Direction', icon: Palette },
  { key: 'panel_generation', label: '4. Panel Generation', icon: ImageIcon },
  { key: 'pdf_composition', label: '5. PDF Storybook', icon: FileCheck },
];

export const JobProgress: React.FC<JobProgressProps> = ({ job }) => {
  const output = job.output as ComicPipelineState | undefined;
  const currentStage = output?.stage;

  const isComic = job.module === 'comic';

  const getStageIndex = (stageKey?: string) => {
    if (!stageKey) return 0;
    const idx = STAGES.findIndex((s) => s.key === stageKey);
    return idx >= 0 ? idx : 0;
  };

  const currentStageIdx = getStageIndex(currentStage);

  return (
    <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-5 backdrop-blur-md space-y-4">
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Sparkles className="h-4 w-4 text-cyan-300 animate-pulse" />
          <span className="text-sm font-bold text-white uppercase tracking-wider">
            Job #{job.id.substring(0, 8)}
          </span>
        </div>
        <Badge status={job.status} />
      </div>

      {job.status === 'processing' && (
        <div className="space-y-3">
          <div className="flex items-center justify-between text-xs text-slate-300">
            <span>Pipeline Execution in Progress…</span>
            <span className="font-mono text-cyan-300 font-semibold">
              {isComic ? `Stage ${currentStageIdx + 1} / 5` : 'Generating DALL-E Portrait'}
            </span>
          </div>

          <div className="h-2 w-full rounded-full bg-slate-900 overflow-hidden border border-white/10">
            <div
              className="h-full bg-gradient-to-r from-cyan-400 to-indigo-500 transition-all duration-500 rounded-full"
              style={{
                width: isComic
                  ? `${Math.max(15, ((currentStageIdx + 1) / 5) * 100)}%`
                  : '65%',
              }}
            />
          </div>

          {isComic && (
            <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 pt-2">
              {STAGES.map((s, idx) => {
                const Icon = s.icon;
                const isDone = job.status === 'completed' || idx < currentStageIdx;
                const isCurrent = job.status === 'processing' && idx === currentStageIdx;

                return (
                  <div
                    key={s.key}
                    className={`flex flex-col items-center p-2 rounded-xl border text-center transition-all ${
                      isDone
                        ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-300'
                        : isCurrent
                        ? 'border-cyan-400/50 bg-cyan-500/15 text-cyan-200 shadow-md shadow-cyan-500/20 animate-pulse'
                        : 'border-white/5 bg-white/5 text-slate-500'
                    }`}
                  >
                    <Icon className={`h-4 w-4 mb-1 ${isCurrent ? 'animate-bounce' : ''}`} />
                    <span className="text-[10px] font-semibold leading-tight">{s.label.split('. ')[1]}</span>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {job.status === 'failed' && (
        <div className="flex items-start gap-2.5 rounded-xl border border-rose-500/30 bg-rose-500/10 p-3 text-xs text-rose-300">
          <AlertCircle className="h-4 w-4 shrink-0 mt-0.5" />
          <div>
            <p className="font-bold uppercase">Generation Failed</p>
            <p className="mt-0.5">{job.errorMessage || job.errorCode || 'An error occurred during pipeline processing.'}</p>
          </div>
        </div>
      )}
    </div>
  );
};
