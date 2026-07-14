import React, { useState } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { requestUploadURL, uploadFileToURL, createComicGeneration, fetchGenerationJob, getApiErrorMessage } from '../../api/client';
import type { GenerationJob } from '../../api/types';
import { FileUploader } from './FileUploader';
import { JobProgress } from './JobProgress';
import { JobResultCard } from './JobResultCard';
import { Input, Textarea, Select } from '../ui/Input';
import { Button } from '../ui/Button';
import { BookOpen, ArrowRight, AlertCircle, Sparkles } from 'lucide-react';

interface ComicFormProps {
  defaultProtagonistName?: string;
  onJobSubmitted?: (jobId: string) => void;
}

export const ComicForm: React.FC<ComicFormProps> = ({ defaultProtagonistName = '', onJobSubmitted }) => {
  const queryClient = useQueryClient();
  const [file, setFile] = useState<File | null>(null);
  const [protagonistName, setProtagonistName] = useState(defaultProtagonistName);
  const [biggestHigh, setBiggestHigh] = useState('');
  const [biggestLow, setBiggestLow] = useState('');
  const [language, setLanguage] = useState('English');
  const [tone, setTone] = useState('inspiring');
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
      createComicGeneration({
        sourceAssetId: payload.assetId,
        idempotencyKey: payload.idempotencyKey,
        biggestHigh,
        biggestLow,
        protagonistName: protagonistName || 'Hero',
        language,
        tone,
      }),
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setErrorMessage(null);

    if (!file) {
      setErrorMessage('Please select a reference selfie image file.');
      return;
    }
    if (!biggestHigh.trim() || !biggestLow.trim()) {
      setErrorMessage('Please describe both your biggest high and biggest low life moments.');
      return;
    }

    try {
      setStatusText('1/3 Requesting asset upload URL…');
      const intent = await uploadIntentMutation.mutateAsync(file);

      setStatusText('2/3 Uploading photo to storage…');
      await uploadFileToURL(intent.uploadUrl, file);

      setStatusText('3/3 Queueing comic storybook pipeline worker…');
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
    setBiggestHigh('');
    setBiggestLow('');
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
            label="1. Protagonist Selfie Reference"
            hint="Used by AI Art Direction agent to maintain consistent character features across panels"
          />

          <div className="space-y-4 pt-2">
            <h4 className="text-xs font-semibold uppercase tracking-wider text-cyan-300">
              2. Narrative Story Prompts
            </h4>

            <Input
              label="Protagonist Display Name"
              value={protagonistName}
              onChange={(e) => setProtagonistName(e.target.value)}
              placeholder="e.g. Alex"
              hint="Name of the main character in the comic storybook"
            />

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Textarea
                label="Your Life's Biggest High (Peak Moment)"
                value={biggestHigh}
                onChange={(e) => setBiggestHigh(e.target.value)}
                placeholder="e.g., Winning the national hackathon championship and getting our first major funding investment."
                hint="Triumphant moment for the narrative arc climax"
              />

              <Textarea
                label="Your Life's Biggest Low (Challenging Moment)"
                value={biggestLow}
                onChange={(e) => setBiggestLow(e.target.value)}
                placeholder="e.g., Struggling during the early startup phase when servers crashed and funds were nearly depleted."
                hint="Obstacle and conflict phase for the narrative arc"
              />
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Select
                label="Story Tone & Visual Vibe"
                value={tone}
                onChange={(e) => setTone(e.target.value)}
                options={[
                  { value: 'inspiring', label: 'Inspiring & Heroic' },
                  { value: 'humorous', label: 'Humorous & Fun' },
                  { value: 'dramatic', label: 'Dramatic & Emotional' },
                  { value: 'cyberpunk', label: 'Futuristic Sci-Fi' },
                  { value: 'fantasy', label: 'Epic Fantasy Myth' },
                ]}
              />

              <Select
                label="Dialogue & Caption Language"
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
                options={[
                  { value: 'English', label: 'English' },
                  { value: 'Hindi', label: 'Hindi' },
                  { value: 'Spanish', label: 'Spanish' },
                  { value: 'French', label: 'French' },
                  { value: 'German', label: 'German' },
                ]}
              />
            </div>
          </div>

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
            disabled={!file || !biggestHigh || !biggestLow}
            leftIcon={<BookOpen className="h-5 w-5" />}
            rightIcon={<ArrowRight className="h-4 w-4" />}
          >
            Generate Personal Comic Storybook (1 Credit)
          </Button>
        </form>
      ) : (
        <div className="space-y-6">
          {job && <JobProgress job={job} />}
          {job && job.status === 'completed' && <JobResultCard job={job} onReset={resetForm} />}
          {job && job.status === 'failed' && (
            <div className="pt-2 text-center">
              <Button variant="outline" onClick={resetForm}>
                Try Another Storybook
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
