import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import type { PublicConfig, CurrentUser, GenerationJob } from '../api/types';
import { fetchGenerationJob } from '../api/client';
import { Shell } from '../components/layout/Shell';
import { JobProgress } from '../components/generation/JobProgress';
import { JobResultCard } from '../components/generation/JobResultCard';
import { Button } from '../components/ui/Button';
import { ArrowLeft, Loader2, AlertCircle } from 'lucide-react';

interface JobDetailPageProps {
  config: PublicConfig;
  user: CurrentUser;
  onLogout: () => void;
}

export const JobDetailPage: React.FC<JobDetailPageProps> = ({ config, user, onLogout }) => {
  const { id } = useParams<{ id: string }>();

  const { data: job, isLoading, isError } = useQuery({
    queryKey: ['generation-job', id],
    queryFn: () => fetchGenerationJob(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const j = query.state.data as GenerationJob | undefined;
      if (!j) return 2000;
      if (j.status === 'completed' || j.status === 'failed') return false;
      return 2000;
    },
  });

  return (
    <Shell user={user} authMode={config.authMode} onLogout={onLogout}>
      <div className="space-y-6 py-6 max-w-4xl mx-auto">
        <Link to="/generations" className="inline-flex items-center gap-1.5 text-xs font-semibold text-cyan-300 hover:underline">
          <ArrowLeft className="h-4 w-4" />
          <span>Back to My Creations</span>
        </Link>

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-slate-300">
            <Loader2 className="h-8 w-8 animate-spin text-cyan-300 mb-3" />
            <p className="text-sm font-semibold">Loading generation details…</p>
          </div>
        ) : isError || !job ? (
          <div className="rounded-3xl border border-rose-500/30 bg-rose-500/10 p-8 text-center space-y-4">
            <AlertCircle className="h-10 w-10 text-rose-400 mx-auto" />
            <h3 className="text-xl font-bold text-white">Generation Job Not Found</h3>
            <p className="text-xs text-slate-300">
              The requested generation job ID does not exist or you do not have permission to view it.
            </p>
            <Link to="/generations">
              <Button variant="outline">Return to Creations</Button>
            </Link>
          </div>
        ) : (
          <div className="space-y-6">
            <JobProgress job={job} />
            {job.status === 'completed' && <JobResultCard job={job} />}
          </div>
        )}
      </div>
    </Shell>
  );
};
