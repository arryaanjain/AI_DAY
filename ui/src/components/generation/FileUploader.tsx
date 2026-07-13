import React, { useState, useRef } from 'react';
import { Upload, CheckCircle2, X } from 'lucide-react';

interface FileUploaderProps {
  onFileSelect: (file: File | null) => void;
  selectedFile: File | null;
  label?: string;
  hint?: string;
}

export const FileUploader: React.FC<FileUploaderProps> = ({
  onFileSelect,
  selectedFile,
  label = 'Source Selfie Photo',
  hint = 'Upload a clear front-facing image (JPEG, PNG, WebP up to 10MB)',
}) => {
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (file: File | null) => {
    if (!file) {
      setPreviewUrl(null);
      onFileSelect(null);
      return;
    }
    const url = URL.createObjectURL(file);
    setPreviewUrl(url);
    onFileSelect(file);
  };

  const clearFile = (e: React.MouseEvent) => {
    e.stopPropagation();
    setPreviewUrl(null);
    onFileSelect(null);
    if (inputRef.current) inputRef.current.value = '';
  };

  return (
    <div className="w-full">
      <label className="block text-xs font-semibold uppercase tracking-wider text-slate-300 mb-2">
        {label}
      </label>

      <div
        onClick={() => inputRef.current?.click()}
        className={`group relative flex flex-col items-center justify-center rounded-2xl border-2 border-dashed p-6 text-center cursor-pointer transition-all duration-200 ${
          selectedFile
            ? 'border-cyan-300/60 bg-cyan-500/5'
            : 'border-white/15 bg-slate-950/60 hover:border-cyan-300/40 hover:bg-slate-950/80'
        }`}
      >
        <input
          ref={inputRef}
          type="file"
          accept="image/jpeg,image/png,image/webp"
          className="hidden"
          onChange={(e) => handleFileChange(e.target.files?.[0] || null)}
        />

        {previewUrl && selectedFile ? (
          <div className="flex flex-col items-center gap-3">
            <div className="relative h-28 w-28 overflow-hidden rounded-2xl border border-cyan-300/40 shadow-lg shadow-cyan-950/50">
              <img
                src={previewUrl}
                alt="Selfie Preview"
                className="h-full w-full object-cover"
              />
              <button
                type="button"
                onClick={clearFile}
                className="absolute top-1 right-1 rounded-full bg-slate-950/80 p-1 text-slate-300 hover:text-white hover:bg-rose-500 transition-colors"
              >
                <X className="h-3.5 w-3.5" />
              </button>
            </div>
            <div className="flex items-center gap-1.5 text-xs font-medium text-cyan-300">
              <CheckCircle2 className="h-4 w-4" />
              <span className="truncate max-w-[200px]">{selectedFile.name}</span>
              <span className="text-slate-400">({(selectedFile.size / (1024 * 1024)).toFixed(2)} MB)</span>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center gap-2">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-cyan-300/10 text-cyan-300 border border-cyan-300/20 group-hover:scale-110 transition-transform">
              <Upload className="h-6 w-6" />
            </div>
            <p className="mt-2 text-sm font-semibold text-white group-hover:text-cyan-300 transition-colors">
              Click to select photo
            </p>
            <p className="text-xs text-slate-400">{hint}</p>
          </div>
        )}
      </div>
    </div>
  );
};
