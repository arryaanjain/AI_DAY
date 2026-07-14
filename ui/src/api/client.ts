import axios from 'axios';
import type {
  PublicConfig,
  CurrentUser,
  CreditBalance,
  UploadIntent,
  PaymentOrder,
  GenerationJob,
  CreateComicRequest,
  CreatePixartRequest,
} from './types';

export const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080',
  withCredentials: true,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('ai_day_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

function unwrap<T>(response: { data: { data: T } }): T {
  return response.data.data;
}

export function formatMoney(paise: number, currency = 'INR'): string {
  if (currency === 'INR') {
    return `₹${(paise / 100).toFixed(0)}`;
  }
  return `${currency} ${(paise / 100).toFixed(2)}`;
}

// Config & User Auth API
export async function fetchPublicConfig(): Promise<PublicConfig> {
  const res = await api.get('/api/v1/config/public');
  return unwrap<PublicConfig>(res);
}

export async function fetchCurrentUser(): Promise<CurrentUser | null> {
  try {
    const res = await api.get('/api/v1/auth/me');
    return unwrap<CurrentUser>(res);
  } catch (error) {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      return null;
    }
    throw error;
  }
}

export async function fetchCreditBalance(): Promise<number> {
  try {
    const res = await api.get('/api/v1/credits/balance');
    const data = unwrap<CreditBalance>(res);
    return data.available;
  } catch {
    return 0;
  }
}

export async function sendPhoneOTP(phone: string): Promise<{ status: string }> {
  const res = await api.post('/api/v1/auth/phone/send-otp', { phone });
  return unwrap<{ status: string }>(res);
}

export async function resendPhoneOTP(phone: string): Promise<{ status: string }> {
  const res = await api.post('/api/v1/auth/phone/resend-otp', { phone });
  return unwrap<{ status: string }>(res);
}

export async function verifyPhoneOTP(phone: string, code: string): Promise<{ userId: string; token?: string }> {
  const res = await api.post('/api/v1/auth/phone/verify-otp', { phone, code });
  const data = unwrap<{ userId: string; token?: string }>(res);
  if (data.token) {
    localStorage.setItem('ai_day_token', data.token);
  }
  return data;
}

export async function logoutUser(): Promise<void> {
  try {
    await api.post('/api/v1/auth/logout');
  } finally {
    localStorage.removeItem('ai_day_token');
  }
}

// Assets & Storage API
export async function requestUploadURL(
  filename: string,
  mimeType: string,
  sizeBytes: number
): Promise<UploadIntent> {
  const res = await api.post('/api/v1/assets/upload-url', {
    assetType: 'source_selfie',
    filename,
    mimeType,
    sizeBytes,
  });
  return unwrap<UploadIntent>(res);
}

export async function uploadFileToURL(uploadUrl: string, file: File): Promise<void> {
  if (uploadUrl.startsWith('http://') || uploadUrl.startsWith('https://')) {
    await fetch(uploadUrl, {
      method: 'PUT',
      headers: {
        'Content-Type': file.type || 'application/octet-stream',
      },
      body: file,
    });
  }
}

// Generations API
export async function createPixartGeneration(payload: CreatePixartRequest): Promise<{ jobId: string; status: string }> {
  const res = await api.post('/api/v1/generations/pixart', payload);
  return unwrap<{ jobId: string; status: string }>(res);
}

export async function createComicGeneration(payload: CreateComicRequest): Promise<{ jobId: string; status: string }> {
  const res = await api.post('/api/v1/generations/comic', payload);
  return unwrap<{ jobId: string; status: string }>(res);
}

export async function fetchGenerationJob(jobId: string): Promise<GenerationJob> {
  const res = await api.get(`/api/v1/generations/${jobId}`);
  return unwrap<GenerationJob>(res);
}

export async function retryGenerationJob(jobId: string): Promise<{ jobId: string; status: string }> {
  const res = await api.post(`/api/v1/generations/${jobId}/retry`);
  return unwrap<{ jobId: string; status: string }>(res);
}

export async function fetchGenerationJobs(): Promise<GenerationJob[]> {
  const res = await api.get('/api/v1/generations');
  return unwrap<GenerationJob[]>(res);
}

// Payments API
export async function createPaymentOrder(module: 'pixel_portrait' | 'comic', quantity = 1): Promise<PaymentOrder> {
  const res = await api.post('/api/v1/payments/orders', { module, quantity });
  return unwrap<PaymentOrder>(res);
}

export async function verifyPaymentOrder(
  razorpayOrderId: string,
  razorpayPaymentId: string,
  razorpaySignature: string
): Promise<{ status: string; message: string }> {
  const res = await api.post('/api/v1/payments/verify', {
    razorpayOrderId,
    razorpayPaymentId,
    razorpaySignature,
  });
  return unwrap<{ status: string; message: string }>(res);
}

// Error Message Extractor
export function getApiErrorMessage(error: unknown): string {
  if (axios.isAxiosError(error)) {
    const data = error.response?.data;
    if (data?.error?.message) {
      return data.error.message;
    }
    if (typeof data === 'string') return data;
  }
  if (error instanceof Error) return error.message;
  return 'An unexpected error occurred. Please try again.';
}
