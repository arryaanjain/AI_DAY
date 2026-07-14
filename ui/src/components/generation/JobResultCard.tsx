import React from 'react';
import type { GenerationJob, ComicPipelineState, PixartJobOutput } from '../../api/types';
import { Button } from '../ui/Button';
import { Download, CheckCircle2, BookOpen, ExternalLink, RefreshCw } from 'lucide-react';

interface JobResultCardProps {
  job: GenerationJob;
  onReset?: () => void;
}

export const JobResultCard: React.FC<JobResultCardProps> = ({ job, onReset }) => {
  const apiBase = import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080';

  if (job.status !== 'completed') return null;

  const isComic = job.module === 'comic';
  const comicOutput = isComic ? (job.output as ComicPipelineState) : undefined;
  const pixartOutput = !isComic ? (job.output as PixartJobOutput) : undefined;

  // Derive asset download URL if available (avoiding file:// scheme URLs)
  const pdfUrl = comicOutput?.pdfAssetId
    ? `${apiBase}/api/v1/assets/${comicOutput.pdfAssetId}/download`
    : (comicOutput?.pdfUrl && !comicOutput.pdfUrl.startsWith('file://')
        ? (comicOutput.pdfUrl.startsWith('/') ? `${apiBase}${comicOutput.pdfUrl}` : comicOutput.pdfUrl)
        : `${apiBase}/api/v1/storage/users/${job.userId}/comic/${job.id}/story.pdf`);

  const downloadUrl = pixartOutput?.assetId
    ? `${apiBase}/api/v1/assets/${pixartOutput.assetId}/download`
    : `${apiBase}/api/v1/storage/users/${job.userId}/pixart/${job.id}/image.png`;

  const imageUrl = pixartOutput?.assetId
    ? `${apiBase}/api/v1/assets/${pixartOutput.assetId}/download`
    : (pixartOutput?.imageUrl && !pixartOutput.imageUrl.startsWith('file://')
        ? (pixartOutput.imageUrl.startsWith('/') ? `${apiBase}${pixartOutput.imageUrl}` : pixartOutput.imageUrl)
        : `${apiBase}/api/v1/storage/users/${job.userId}/pixart/${job.id}/image.png`);

  return (
    <div className="rounded-[2rem] border border-cyan-300/30 bg-gradient-to-b from-cyan-950/20 to-slate-900/60 p-6 backdrop-blur-xl space-y-6 shadow-2xl shadow-cyan-950/40 animate-in fade-in zoom-in-95 duration-300">
      <div className="flex items-center justify-between border-b border-white/10 pb-4">
        <div className="flex items-center gap-2.5">
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-emerald-500/20 text-emerald-300 border border-emerald-500/30">
            <CheckCircle2 className="h-5 w-5" />
          </div>
          <div>
            <h3 className="text-lg font-bold text-white">
              {isComic ? 'Comic Storybook Ready!' : 'PixArt Stylized Portrait Ready!'}
            </h3>
            <p className="text-xs text-slate-400">Generation completed successfully</p>
          </div>
        </div>

        {onReset && (
          <Button variant="outline" size="sm" onClick={onReset} leftIcon={<RefreshCw className="h-3.5 w-3.5" />}>
            New Generation
          </Button>
        )}
      </div>

      {!isComic ? (
        // PixArt Output Display
        <div className="flex flex-col items-center space-y-4">
          <div className="relative overflow-hidden rounded-2xl border-2 border-cyan-300/40 shadow-2xl shadow-cyan-500/20 max-w-md w-full aspect-square bg-slate-950">
            <img
              src={imageUrl}
              alt="Generated PixArt Portrait"
              className="h-full w-full object-cover transition-transform duration-500 hover:scale-105"
              onError={(e) => {
                // Fallback demo image if storage server is in noop mode
                (e.target as HTMLImageElement).src = 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&w=800&q=80';
              }}
            />
          </div>
          <div className="flex flex-wrap gap-3 pt-2">
            <a href={downloadUrl} target="_blank" rel="noopener noreferrer" download={`pixart-${job.id}.png`}>
              <Button size="md" leftIcon={<Download className="h-4 w-4" />}>
                Download HD Image
              </Button>
            </a>
          </div>
        </div>
      ) : (
        // Comic Output Display
        <div className="space-y-5">
          {comicOutput?.narrative && (
            <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-5 space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="text-base font-bold text-cyan-300 flex items-center gap-2">
                  <BookOpen className="h-4 w-4" />
                  <span>{comicOutput.narrative.title}</span>
                </h4>
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 bg-white/5 px-2.5 py-1 rounded-full">
                  Theme: {comicOutput.narrative.theme}
                </span>
              </div>
              <p className="text-xs text-slate-300 leading-relaxed italic">
                "{comicOutput.narrative.logline}"
              </p>

              {/* Panels Grid Preview */}
              {comicOutput.panels && comicOutput.panels.length > 0 && (
                <div className="mt-4 pt-4 border-t border-white/10">
                  <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3">
                    Generated Comic Panels ({comicOutput.panels.length})
                  </p>
                  <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
                    {comicOutput.panels.map((panel) => (
                      <div key={panel.panelNumber} className="relative aspect-square overflow-hidden rounded-xl border border-white/10 bg-slate-900 group">
                        <img
                          src={`${apiBase}/storage/${panel.objectKey}`}
                          alt={`Panel ${panel.panelNumber}`}
                          className="h-full w-full object-cover transition-transform group-hover:scale-105"
                          onError={(e) => {
                            (e.target as HTMLImageElement).src = `https://picsum.photos/seed/${panel.panelNumber}/400/400`;
                          }}
                        />
                        <div className="absolute bottom-2 left-2 rounded-md bg-slate-950/80 px-2 py-0.5 text-[10px] font-bold text-cyan-300">
                          Panel #{panel.panelNumber}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          <div className="flex flex-wrap items-center justify-between gap-4 pt-2">
            <div className="text-xs text-slate-400">
              PDF Storybook formatted with print-ready resolution and dialogue captions.
            </div>
            <a href={pdfUrl || '#'} target="_blank" rel="noopener noreferrer" download={`comic-storybook-${job.id}.pdf`}>
              <Button size="lg" leftIcon={<Download className="h-5 w-5" />} rightIcon={<ExternalLink className="h-4 w-4" />}>
                Download PDF Storybook
              </Button>
            </a>
          </div>
        </div>
      )}
    </div>
  );
};
