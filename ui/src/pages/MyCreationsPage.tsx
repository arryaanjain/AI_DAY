import React, { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import type { PublicConfig, CurrentUser, GenerationJob } from '../api/types';
import { fetchGenerationJobs } from '../api/client';
import { Shell } from '../components/layout/Shell';
import { GenerationCard } from '../components/generation/GenerationCard';
import { JobResultCard } from '../components/generation/JobResultCard';
import { JobProgress } from '../components/generation/JobProgress';
import { Modal } from '../components/ui/Modal';
import { Button } from '../components/ui/Button';
import { Layers, Sparkles, BookOpen, Plus, RefreshCw, Loader2, Image as ImageIcon } from 'lucide-react';
import { Link } from 'react-router-dom';

interface MyCreationsPageProps {
  config: PublicConfig;
  user: CurrentUser;
  onLogout: () => void;
}

export const MyCreationsPage: React.FC<MyCreationsPageProps> = ({ config, user, onLogout }) => {
  const [activeFilter, setActiveFilter] = useState<'all' | 'pixart' | 'comic'>('all');
  const [selectedJob, setSelectedJob] = useState<GenerationJob | null>(null);

  const { data: jobs = [], isLoading, refetch, isFetching } = useQuery({
    queryKey: ['user-generations'],
    queryFn: fetchGenerationJobs,
    refetchInterval: 5000,
  });

  const filteredJobs = jobs.filter((job) => {
    if (activeFilter === 'pixart') return job.module === 'pixel_portrait';
    if (activeFilter === 'comic') return job.module === 'comic';
    return true;
  });

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
              <GenerationCard key={job.id} job={job} onSelect={(j) => setSelectedJob(j)} />
            ))}
          </div>
        )}

        {/* Detail Modal */}
        {selectedJob && (
          <Modal isOpen={!!selectedJob} onClose={() => setSelectedJob(null)} title={`Job #${selectedJob.id.substring(0, 8)}`} maxWidth="xl">
            <div className="space-y-6 pt-2">
              <JobProgress job={selectedJob} />
              {selectedJob.status === 'completed' && <JobResultCard job={selectedJob} />}
            </div>
          </Modal>
        )}
      </div>
    </Shell>
  );
};
