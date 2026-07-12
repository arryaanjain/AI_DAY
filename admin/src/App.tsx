import { type FormEvent, type ReactNode, useEffect, useMemo, useState } from 'react'
import { Link, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { LayoutDashboard, Users, CreditCard, Sparkles, Webhook, ScrollText, ShieldAlert, LogOut } from 'lucide-react'
import axios from 'axios'

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080', withCredentials: true })

type PublicConfig = { authMode: 'phone' | 'microsoft'; pricePerGenerationPaise: number; currency: string; enabledModules: string[] }
type CurrentUser = { id: string; email: string | null; phone: string | null; displayName: string; isAdmin: boolean }
type Stats = { totalUsers: number; completedJobs: number; totalAssets: number }
type UserRow = { id: string; email: string; phone: string; displayName: string; status: string; createdAt: string }
type GenerationRow = { id: string; userId: string; module: string; status: string; createdAt: string }

const navItems = [
    [LayoutDashboard, 'Dashboard', '/dashboard'],
    [Users, 'Users', '/users'],
    [Sparkles, 'Generations', '/generations'],
    [CreditCard, 'Payments', '/payments'],
    [Webhook, 'Webhooks', '/webhooks'],
    [ScrollText, 'Audit logs', '/audit-logs'],
] as const

function unwrap<T>(response: { data: { data: T } }) { return response.data.data }
function usePublicConfig() { return useQuery({ queryKey: ['admin-public-config'], queryFn: async () => unwrap<PublicConfig>(await api.get('/api/v1/config/public')) }) }
function useCurrentUser() { return useQuery({ queryKey: ['admin-me'], queryFn: async () => { try { return unwrap<CurrentUser>(await api.get('/api/v1/auth/me')) } catch (error) { if (axios.isAxiosError(error) && error.response?.status === 401) return null; throw error } } }) }

function Shell({ children, user, authMode, onLogout }: { children: ReactNode; user: CurrentUser | null; authMode: PublicConfig['authMode']; onLogout: () => void }) {
    return <div className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(34,211,238,0.14),_transparent_35%),linear-gradient(180deg,#020617_0%,#0f172a_45%,#020617_100%)] text-slate-100 md:flex"><aside className="border-b border-white/10 bg-white/5 p-5 backdrop-blur-xl md:min-h-screen md:w-72 md:border-b-0 md:border-r"><Link to="/dashboard" className="text-xl font-bold tracking-tight">AI Day Admin</Link><p className="mt-2 text-xs uppercase tracking-[0.25em] text-cyan-300">{authMode === 'phone' ? 'SMS admin session' : 'Email admin session'}</p><nav className="mt-8 grid gap-2">{navItems.map(([Icon, label, to]) => <Link key={to} to={to} className="flex items-center gap-3 rounded-2xl border border-white/5 bg-slate-950/40 px-3 py-3 text-sm text-slate-300 transition hover:border-cyan-300/40 hover:bg-white/10 hover:text-white"><Icon size={18} />{label}</Link>)}</nav><div className="mt-8 rounded-2xl border border-white/10 bg-slate-950/50 p-4 text-sm text-slate-300"><p className="font-semibold text-white">Session</p><p className="mt-1">{user ? user.displayName : 'No admin session'}</p><p className="mt-1 truncate">{user?.email || user?.phone || user?.id || 'Sign in with the main app first.'}</p>{user && <button onClick={onLogout} className="mt-4 inline-flex items-center gap-2 rounded-full bg-cyan-300 px-4 py-2 font-semibold text-slate-950"><LogOut size={16} />Logout</button>}</div></aside><main className="flex-1 p-6 md:p-8">{children}</main></div>
}

function StatCard({ label, value, hint }: { label: string; value: string; hint: string }) { return <article className="rounded-[1.75rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl"><p className="text-sm uppercase tracking-[0.2em] text-slate-400">{label}</p><p className="mt-3 text-3xl font-semibold">{value}</p><p className="mt-2 text-sm text-slate-300">{hint}</p></article> }

function TableShell({ title, subtitle, children }: { title: string; subtitle: string; children: ReactNode }) { return <section className="rounded-[2rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl"><div className="flex items-end justify-between gap-4"><div><h1 className="text-4xl font-bold">{title}</h1><p className="mt-2 text-slate-300">{subtitle}</p></div></div><div className="mt-6">{children}</div></section> }

function Login({ config, onLoggedIn }: { config: PublicConfig; onLoggedIn: () => void }) {
    const location = useLocation()
    const navigate = useNavigate()
    const queryClient = useQueryClient()
    const next = new URLSearchParams(location.search).get('next') || '/dashboard'
    const [phone, setPhone] = useState('+91')
    const [code, setCode] = useState('')
    const [email, setEmail] = useState('')
    const [sent, setSent] = useState(false)
    const sendOtp = useMutation({ mutationFn: async () => unwrap<{ status: string }>(await api.post('/api/v1/auth/phone/send-otp', { phone })) })
    const verifyOtp = useMutation({ mutationFn: async () => unwrap<{ userId: string }>(await api.post('/api/v1/auth/phone/verify-otp', { phone, code })) })

    function continueWithEmail() {
        const params = new URLSearchParams({ next })
        if (email.trim()) params.set('email', email.trim())
        window.location.href = `${api.defaults.baseURL}/api/v1/auth/microsoft/start?${params.toString()}`
    }

    return <div className="grid min-h-screen place-items-center px-6 text-slate-100"><div className="w-full max-w-md rounded-[2rem] border border-white/10 bg-white/5 p-8 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.25em] text-cyan-300">Admin sign in</p><h1 className="mt-4 text-3xl font-bold">{config.authMode === 'phone' ? 'Approve with SMS' : 'Approve with Email'}</h1><p className="mt-3 text-sm text-slate-300">Use the same backend auth flow as the main app, then the admin routes will verify <span className="font-semibold text-white">isAdmin</span>.</p>{config.authMode === 'phone' ? <form className="mt-8 space-y-4" onSubmit={async (event: FormEvent) => { event.preventDefault(); if (!sent) { await sendOtp.mutateAsync(); setSent(true); return } await verifyOtp.mutateAsync(); await queryClient.refetchQueries({ queryKey: ['admin-me'] }); onLoggedIn(); navigate(next, { replace: true }) }}><label className="block text-sm font-medium text-slate-200">Phone number<input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="+919876543210" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label>{sent && <label className="block text-sm font-medium text-slate-200">OTP code<input value={code} onChange={(event) => setCode(event.target.value)} placeholder="123456" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label>}<button type="submit" className="w-full rounded-2xl bg-cyan-300 px-4 py-3 font-semibold text-slate-950">{sent ? 'Verify code' : 'Send OTP'}</button></form> : <div className="mt-8 space-y-4"><label className="block text-sm font-medium text-slate-200">Email address<input value={email} onChange={(event) => setEmail(event.target.value)} placeholder="admin@example.com" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><button type="button" onClick={continueWithEmail} className="w-full rounded-2xl bg-cyan-300 px-4 py-3 font-semibold text-slate-950">Continue with Email</button></div>}</div></div>
}

function Dashboard() {
    const statsQuery = useQuery({ queryKey: ['admin-stats'], queryFn: async () => unwrap<{ stats: Stats }>(await api.get('/api/v1/admin/stats')) })
    const stats = statsQuery.data?.stats
    return <section className="space-y-6"><TableShell title="Dashboard" subtitle="Live backend counters from the admin endpoints."><div className="grid gap-4 md:grid-cols-3"><StatCard label="Total users" value={String(stats?.totalUsers ?? '—')} hint="Counts rows in users" /><StatCard label="Completed jobs" value={String(stats?.completedJobs ?? '—')} hint="Counts finished generation jobs" /><StatCard label="Assets" value={String(stats?.totalAssets ?? '—')} hint="Counts uploaded assets" /></div></TableShell></section>
}

function UsersPage() {
    const usersQuery = useQuery({ queryKey: ['admin-users'], queryFn: async () => unwrap<{ users: UserRow[]; total: number }>(await api.get('/api/v1/admin/users')) })
    return <TableShell title="Users" subtitle="Most recent users from the backend users table."><div className="overflow-x-auto"><table className="min-w-full text-left text-sm"><thead className="text-slate-400"><tr><th className="py-3 pr-4">ID</th><th className="py-3 pr-4">Email</th><th className="py-3 pr-4">Phone</th><th className="py-3 pr-4">Name</th><th className="py-3 pr-4">Status</th><th className="py-3 pr-4">Created</th></tr></thead><tbody>{usersQuery.data?.users.map((user) => <tr key={user.id} className="border-t border-white/5"><td className="py-3 pr-4 font-mono text-xs text-slate-300">{user.id}</td><td className="py-3 pr-4">{user.email || '—'}</td><td className="py-3 pr-4">{user.phone || '—'}</td><td className="py-3 pr-4">{user.displayName || '—'}</td><td className="py-3 pr-4">{user.status}</td><td className="py-3 pr-4 text-slate-400">{user.createdAt}</td></tr>)}</tbody></table></div></TableShell>
}

function GenerationsPage() {
    const generationsQuery = useQuery({ queryKey: ['admin-generations'], queryFn: async () => unwrap<{ jobs: GenerationRow[]; total: number }>(await api.get('/api/v1/admin/generations')) })
    return <TableShell title="Generations" subtitle="Queued and completed generation jobs from the backend."><div className="overflow-x-auto"><table className="min-w-full text-left text-sm"><thead className="text-slate-400"><tr><th className="py-3 pr-4">ID</th><th className="py-3 pr-4">User</th><th className="py-3 pr-4">Module</th><th className="py-3 pr-4">Status</th><th className="py-3 pr-4">Created</th></tr></thead><tbody>{generationsQuery.data?.jobs.map((job) => <tr key={job.id} className="border-t border-white/5"><td className="py-3 pr-4 font-mono text-xs text-slate-300">{job.id}</td><td className="py-3 pr-4 font-mono text-xs text-slate-300">{job.userId}</td><td className="py-3 pr-4">{job.module}</td><td className="py-3 pr-4">{job.status}</td><td className="py-3 pr-4 text-slate-400">{job.createdAt}</td></tr>)}</tbody></table></div></TableShell>
}

function PlaceholderPage({ title }: { title: string }) { return <TableShell title={title} subtitle="This backend route is not wired yet, but the navigation is in place for the next backend phase."><div className="flex items-center gap-3 rounded-2xl border border-amber-400/20 bg-amber-400/10 p-4 text-sm text-amber-100"><ShieldAlert size={18} />The backend currently exposes dashboard, users, and generations. Add the route first, then wire this page to it.</div></TableShell> }

export default function App() {
    const configQuery = usePublicConfig()
    const userQuery = useCurrentUser()
    const queryClient = useQueryClient()
    const navigate = useNavigate()
    const config = configQuery.data
    const user = userQuery.data ?? null

    const logout = async () => {
        await api.post('/api/v1/auth/logout')
        await queryClient.invalidateQueries({ queryKey: ['admin-me'] })
        navigate('/login', { replace: true })
    }

    const adminReady = useMemo(() => !!user?.isAdmin, [user])

    useEffect(() => {
        if (!configQuery.isLoading && config && !user) {
            navigate('/login', { replace: true })
        }
    }, [config, configQuery.isLoading, navigate, user])

    if (configQuery.isLoading || !config) {
        return <main className="grid min-h-screen place-items-center bg-slate-950 text-slate-300">Loading admin panel…</main>
    }

    if (userQuery.isLoading) {
        return <main className="grid min-h-screen place-items-center bg-slate-950 text-slate-300">Loading session…</main>
    }

    if (!user) {
        return <Routes><Route path="/login" element={<Login config={config} onLoggedIn={() => void 0} />} /><Route path="*" element={<Navigate to="/login" replace />} /></Routes>
    }

    if (!adminReady) {
        return <div className="grid min-h-screen place-items-center bg-slate-950 px-6 text-slate-100"><div className="w-full max-w-lg rounded-[2rem] border border-white/10 bg-white/5 p-8 text-center backdrop-blur-xl"><ShieldAlert className="mx-auto text-amber-300" size={40} /><h1 className="mt-4 text-3xl font-bold">Admin access required</h1><p className="mt-3 text-slate-300">The backend session is valid, but this account does not have <span className="font-semibold text-white">isAdmin</span> set. Use an admin user and log in again.</p><button onClick={logout} className="mt-6 rounded-2xl bg-cyan-300 px-5 py-3 font-semibold text-slate-950">Return to sign in</button></div></div>
    }

    return <Shell user={user} authMode={config.authMode} onLogout={logout}><Routes><Route path="/dashboard" element={<Dashboard />} /><Route path="/users" element={<UsersPage />} /><Route path="/generations" element={<GenerationsPage />} /><Route path="/payments" element={<PlaceholderPage title="Payments" />} /><Route path="/webhooks" element={<PlaceholderPage title="Webhooks" />} /><Route path="/audit-logs" element={<PlaceholderPage title="Audit logs" />} /><Route path="*" element={<Navigate to="/dashboard" replace />} /></Routes></Shell>
}
