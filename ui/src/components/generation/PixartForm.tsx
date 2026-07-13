import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { requestUploadURL, uploadFileToURL, createPixartGeneration, fetchGenerationJob, getApiErrorMessage } from '../../api/client';
import type { GenerationJob } from '../../api/types';
import { FileUploader } from './FileUploader';
import { JobProgress } from './JobProgress';
import { JobResultCard } from './JobResultCard';
import { Button } from '../ui/Button';
import { Sparkles, ArrowRight, AlertCircle } from 'lucide-react';

interface PixartFormProps {
  onJobSubmitted?: (jobId: string) => void;
}

export const PixartForm: React.FC<PixartFormProps> = ({ onJobSubmitted }) => {
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [activeJobId, setActiveJobId] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [statusText, setStatusText] = useState<string | null>(null);

  // Poll active generation job status
  const { data: job } = useQuery({
    queryKey: ['generation-job', activeJobId],
    queryFn: () => fetchGenerationJob(activeJobId!),
    enabled: !!activeJobId,
    refetchInterval: (query) => {
      const j = query.state.data as GenerationJob | undefined;
      if (!j) return 2000;
      if (j.status === 'completed' || j.status === 'failed') return false;
      return 2000;
    },
  });

  const uploadIntentMutation = useMutation({
    mutationFn: (f: File) => requestUploadURL(f.name, f.type || 'image/jpeg', f.size),
  });

  const createJobMutation = useMutation({
    mutationFn: (payload: { assetId: string; idempotencyKey: string }) =>
      createPixartGeneration({ sourceAssetId: payload.assetId, idempotencyKey: payload.idempotencyKey }),
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!file) {
      setErrorMessage('Please select a selfie image file first.');
      return;
    }

    try {
      setStatusText('1/3 Requesting asset upload URL…');
      const intent = await uploadIntentMutation.mutateAsync(file);

      setStatusText('2/3 Uploading image to storage…');
      await uploadFileToURL(intent.uploadUrl, file);

      setStatusText('3/3 Queueing PixArt portrait generation…');
      const idempotencyKey = crypto.randomUUID();
      const res = await createJobMutation.mutateAsync({
        assetId: intent.assetId,
        idempotencyKey,
      });

      setActiveJobId(res.jobId);
      setStatusText(null);

      // Invalidate credit balance and user generations query
      await queryClient.invalidateQueries({ queryKey: ['credit-balance'] });
      await queryClient.invalidateQueries({ queryKey: ['user-generations'] });

      if (onJobSubmitted) onJobSubmitted(res.jobId);
    } catch (err) {
      setStatusText(null);
      setErrorMessage(getApiErrorMessage(err));
    }
  };

  const resetForm = () => {
    setFile(null);
    setActiveJobId(null);
    setErrorMessage(null);
    setStatusText(null);
  };

  const isSubmitting = uploadIntentMutation.isPending || createJobMutation.isPending || !!statusText;

  return (
    <div className="space-y-6">
      {!activeJobId ? (
        <form onSubmit={handleSubmit} className="space-y-6">
          <FileUploader
            selectedFile={file}
            onFileSelect={setFile}
            label="1. Upload Front-Facing Selfie"
            hint="Upload a portrait photo to generate 3D PixArt character rendering"
          />

          {errorMessage && (
            <div className="flex items-center gap-2 rounded-2xl border border-rose-500/30 bg-rose-500/10 p-4 text-xs font-medium text-rose-300">
              <AlertCircle className="h-4 w-4 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          )}

          {statusText && (
            <div className="text-xs font-semibold text-cyan-300 animate-pulse flex items-center gap-2">
              <Sparkles className="h-4 w-4" />
              <span>{statusText}</span>
            </div>
          )}

          <Button
            type="submit"
            size="lg"
            className="w-full"
            isLoading={isSubmitting}
            disabled={!file}
            leftIcon={<Sparkles className="h-5 w-5" />}
            rightIcon={<ArrowRight className="h-4 w-4" />}
          >
            Generate PixArt 3D Portrait (1 Credit)
          </Button>
        </form>
      ) : (
        <div className="space-y-6">
          {job && <JobProgress job={job} />}
          {job && job.status === 'completed' && <JobResultCard job={job} onReset={resetForm} />}
          {job && job.status === 'failed' && (
            <div className="pt-2 text-center">
              <Button variant="outline" onClick={resetForm}>
                Try Another Generation
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
