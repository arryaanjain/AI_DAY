import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { PublicConfig, CurrentUser, GenerationJob } from '../api/types';
import { fetchGenerationJobs, retryGenerationJob, getApiErrorMessage } from '../api/client';
import { Shell } from '../components/layout/Shell';
import { GenerationCard } from '../components/generation/GenerationCard';
import { JobResultCard } from '../components/generation/JobResultCard';
import { JobProgress } from '../components/generation/JobProgress';
import { Modal } from '../components/ui/Modal';
import { Button } from '../components/ui/Button';
import { Layers, Sparkles, BookOpen, Plus, RefreshCw, Loader2, Image as ImageIcon, AlertCircle } from 'lucide-react';
import { Link } from 'react-router-dom';

interface MyCreationsPageProps {
  config: PublicConfig;
  user: CurrentUser;
  onLogout: () => void;
}

export const MyCreationsPage: React.FC<MyCreationsPageProps> = ({ config, user, onLogout }) => {
  const queryClient = useQueryClient();
  const [activeFilter, setActiveFilter] = useState<'all' | 'pixart' | 'comic'>('all');
  const [selectedJob, setSelectedJob] = useState<GenerationJob | null>(null);
  const [retryError, setRetryError] = useState<string | null>(null);

  const { data: jobs = [], isLoading, refetch, isFetching } = useQuery({
    queryKey: ['user-generations'],
    queryFn: fetchGenerationJobs,
    refetchInterval: 5000,
  });

  const retryMutation = useMutation({
    mutationFn: (jobId: string) => retryGenerationJob(jobId),
    onSuccess: (_data, jobId) => {
      setRetryError(null);
      // Invalidate both the list and the specific job so polling resumes
      queryClient.invalidateQueries({ queryKey: ['user-generations'] });
      queryClient.invalidateQueries({ queryKey: ['generation-job', jobId] });
      // Update the selected job optimistically so the modal shows 'queued'
      setSelectedJob((prev) => prev ? { ...prev, status: 'queued' } : null);
    },
    onError: (err) => {
      setRetryError(getApiErrorMessage(err));
    },
  });

  const filteredJobs = jobs.filter((job) => {
    if (activeFilter === 'pixart') return job.module === 'pixel_portrait';
    if (activeFilter === 'comic') return job.module === 'comic';
    return true;
  });

  // Keep modal in sync with refreshed job list
  const liveSelectedJob = selectedJob
    ? (jobs.find((j) => j.id === selectedJob.id) ?? selectedJob)
    : null;

  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      <div className="space-y-8 py-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-black text-white flex items-center gap-3">
              <Layers className="h-8 w-8 text-purple-300" />
              <span>My AI Creations</span>
            </h1>
            <p className="mt-1 text-sm text-slate-300">
              View your history of generated 3D PixArt portraits and personal comic storybooks.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <Button
              variant="secondary"
              size="sm"
              onClick={() => refetch()}
              isLoading={isFetching}
              leftIcon={<RefreshCw className="h-3.5 w-3.5" />}
            >
              Refresh
            </Button>

            <Link to="/comic">
              <Button size="sm" leftIcon={<Plus className="h-4 w-4" />}>
                New Storybook
              </Button>
            </Link>
          </div>
        </div>

        {/* Filter Tabs */}
        <div className="flex items-center gap-2 border-b border-white/10 pb-4">
          <button
            onClick={() => setActiveFilter('all')}
            className={`px-4 py-2 rounded-full text-xs font-bold transition-colors ${
              activeFilter === 'all'
                ? 'bg-white/15 text-white border border-white/10'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            All Creations ({jobs.length})
          </button>

          <button
            onClick={() => setActiveFilter('pixart')}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-full text-xs font-bold transition-colors ${
              activeFilter === 'pixart'
                ? 'bg-cyan-500/20 text-cyan-300 border border-cyan-500/30'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            <Sparkles className="h-3.5 w-3.5" />
            <span>PixArt 3D ({jobs.filter((j) => j.module === 'pixel_portrait').length})</span>
          </button>

          <button
            onClick={() => setActiveFilter('comic')}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-full text-xs font-bold transition-colors ${
              activeFilter === 'comic'
                ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
                : 'text-slate-400 hover:text-white hover:bg-white/5'
            }`}
          >
            <BookOpen className="h-3.5 w-3.5" />
            <span>Comic Storybooks ({jobs.filter((j) => j.module === 'comic').length})</span>
          </button>
        </div>

        {/* Jobs Grid */}
        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-slate-400">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-300 mb-3" />
            <p className="text-sm font-semibold">Fetching your creations…</p>
          </div>
        ) : filteredJobs.length === 0 ? (
          <div className="rounded-[2.5rem] border border-white/10 bg-slate-900/40 p-12 text-center space-y-4">
            <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-3xl bg-white/5 text-slate-400 border border-white/10">
              <ImageIcon className="h-8 w-8" />
            </div>
            <h3 className="text-xl font-bold text-white">No creations found</h3>
            <p className="text-sm text-slate-300 max-w-sm mx-auto">
              You haven't submitted any generation jobs yet. Start by creating a PixArt portrait or a personal comic storybook.
            </p>
            <div className="flex items-center justify-center gap-3 pt-2">
              <Link to="/pixart">
                <Button variant="outline" size="sm" leftIcon={<Sparkles className="h-4 w-4" />}>
                  Try PixArt
                </Button>
              </Link>
              <Link to="/comic">
                <Button size="sm" leftIcon={<BookOpen className="h-4 w-4" />}>
                  Create Comic
                </Button>
              </Link>
            </div>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
            {filteredJobs.map((job) => (
              <GenerationCard key={job.id} job={job} onSelect={(j) => { setSelectedJob(j); setRetryError(null); }} />
            ))}
          </div>
        )}

        {/* Detail Modal */}
        {liveSelectedJob && (
          <Modal isOpen={!!liveSelectedJob} onClose={() => { setSelectedJob(null); setRetryError(null); }} title={`Job #${liveSelectedJob.id.substring(0, 8)}`} maxWidth="xl">
            <div className="space-y-6 pt-2">
              <JobProgress job={liveSelectedJob} />
              {liveSelectedJob.status === 'completed' && <JobResultCard job={liveSelectedJob} />}
              {liveSelectedJob.status === 'failed' && (
                <div className="space-y-3">
                  {retryError && (
                    <div className="flex items-center gap-2 rounded-2xl border border-rose-500/30 bg-rose-500/10 p-3 text-xs font-medium text-rose-300">
                      <AlertCircle className="h-4 w-4 shrink-0" />
                      <span>{retryError}</span>
                    </div>
                  )}
                  <div className="flex flex-wrap items-center justify-center gap-3 pt-1">
                    <Button
                      leftIcon={<RefreshCw className="h-4 w-4" />}
                      isLoading={retryMutation.isPending}
                      onClick={() => retryMutation.mutate(liveSelectedJob.id)}
                    >
                      Retry Execution
                    </Button>
                    <Link to="/comic">
                      <Button variant="outline">
                        New Storybook
                      </Button>
                    </Link>
                  </div>
                </div>
              )}
            </div>
          </Modal>
        )}
      </div>
    </Shell>
  );
};

