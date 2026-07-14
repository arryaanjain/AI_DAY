import React from 'react';
import { Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { fetchPublicConfig, fetchCurrentUser, logoutUser } from './api/client';
import type { CurrentUser } from './api/types';
import { LandingPage } from './pages/LandingPage';
import { LoginPage } from './pages/LoginPage';
import { PixartPage } from './pages/PixartPage';
import { ComicPage } from './pages/ComicPage';
import { MyCreationsPage } from './pages/MyCreationsPage';
import { JobDetailPage } from './pages/JobDetailPage';
import { PaymentsPage } from './pages/PaymentsPage';

function RequireAuth({ user, next, children }: { user: CurrentUser | null; next: string; children: React.ReactNode }) {
  if (!user) {
    return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />;
  }
  return <>{children}</>;
}

export default function App() {
  const queryClient = useQueryClient();
  const navigate = useNavigate();

  const configQuery = useQuery({
    queryKey: ['public-config'],
    queryFn: fetchPublicConfig,
  });

  const userQuery = useQuery({
    queryKey: ['me'],
    queryFn: fetchCurrentUser,
  });

  const config = configQuery.data;
  const user = userQuery.data ?? null;

  const handleLogout = async () => {
    try {
      await logoutUser();
    } catch (err) {
      console.error('Logout error:', err);
    }
    await queryClient.invalidateQueries({ queryKey: ['me'] });
    await queryClient.invalidateQueries({ queryKey: ['credit-balance'] });
    navigate('/login', { replace: true });
  };

  const isLoading = configQuery.isLoading || configQuery.isFetching;

  if (isLoading || !config) {
    return (
      <main className="min-h-screen flex items-center justify-center bg-slate-950 text-slate-100 font-sans">
        <div className="flex flex-col items-center gap-3">
          <div className="h-8 w-8 animate-spin rounded-full border-4 border-cyan-300 border-t-transparent" />
          <p className="text-sm font-bold tracking-wider text-slate-300 uppercase">
            Loading AI Day Application…
          </p>
        </div>
      </main>
    );
  }

  return (
    <Routes>
      <Route path="/" element={<LandingPage config={config} user={user} onLogout={handleLogout} />} />
      <Route path="/login" element={<LoginPage config={config} user={user} onLogout={handleLogout} />} />

      <Route
        path="/pixart"
        element={
          <RequireAuth user={user} next="/pixart">
            <PixartPage config={config} user={user!} onLogout={handleLogout} />
          </RequireAuth>
        }
      />

      <Route
        path="/comic"
        element={
          <RequireAuth user={user} next="/comic">
            <ComicPage config={config} user={user!} onLogout={handleLogout} />
          </RequireAuth>
        }
      />

      <Route
        path="/generations"
        element={
          <RequireAuth user={user} next="/generations">
            <MyCreationsPage config={config} user={user!} onLogout={handleLogout} />
          </RequireAuth>
        }
      />

      <Route
        path="/generations/:id"
        element={
          <RequireAuth user={user} next="/generations">
            <JobDetailPage config={config} user={user!} onLogout={handleLogout} />
          </RequireAuth>
        }
      />

      <Route
        path="/payments"
        element={
          <RequireAuth user={user} next="/payments">
            <PaymentsPage config={config} user={user!} onLogout={handleLogout} />
          </RequireAuth>
        }
      />

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
