export type AuthMode = 'phone' | 'microsoft';

export interface PublicConfig {
  authMode: AuthMode;
  pricePerGenerationPaise: number;
  currency: string;
  enabledModules: string[];
}

export interface CurrentUser {
  id: string;
  email: string | null;
  phone: string | null;
  displayName: string;
  isAdmin: boolean;
}

export interface CreditBalance {
  available: number;
}

export interface UploadIntent {
  assetId: string;
  uploadUrl: string;
}

export interface PaymentOrder {
  orderId: string;
  razorpayOrderId: string;
  razorpayKeyId: string;
  amountPaise: number;
  currency: string;
}

export type JobStatus = 'created' | 'queued' | 'processing' | 'completed' | 'failed';

export interface GeneratedPanel {
  panelNumber: number;
  assetId: string;
  objectKey: string;
}

export interface ComicPanelNarrative {
  panelNumber: number;
  visual: string;
  caption: string;
  dialogue?: string[];
  emotion?: string;
}

export interface ComicPageNarrative {
  pageNumber: number;
  purpose: string;
  panels: ComicPanelNarrative[];
}

export interface ComicNarrative {
  title: string;
  theme: string;
  logline: string;
  pages: ComicPageNarrative[];
}

export interface ComicPipelineState {
  stage: 'narrative_planner' | 'safety_review' | 'art_direction' | 'panel_generation' | 'pdf_composition';
  narrative?: ComicNarrative;
  safety?: { safe: boolean; reason: string };
  panels?: GeneratedPanel[];
  pdfAssetId?: string;
  pdfUrl?: string;
}

export interface PixartJobOutput {
  imageUrl?: string;
  assetId?: string;
  providerRequestId?: string;
}

export interface GenerationJob {
  id: string;
  userId: string;
  module: 'pixel_portrait' | 'comic';
  status: JobStatus;
  sourceAssetId?: string;
  input: {
    sourceAssetId?: string;
    idempotencyKey?: string;
    biggestHigh?: string;
    biggestLow?: string;
    protagonistName?: string;
    language?: string;
    tone?: string;
  };
  output?: ComicPipelineState | PixartJobOutput;
  errorCode?: string;
  errorMessage?: string;
  startedAt?: string;
  completedAt?: string;
  createdAt: string;
}

export interface CreateComicRequest {
  sourceAssetId: string;
  idempotencyKey: string;
  biggestHigh: string;
  biggestLow: string;
  protagonistName: string;
  language: string;
  tone: string;
}

export interface CreatePixartRequest {
  sourceAssetId: string;
  idempotencyKey: string;
}
