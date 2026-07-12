import { type FormEvent, type ReactNode, useEffect, useState } from 'react'
import { Link, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import axios from 'axios'

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL || 'http://127.0.0.1:8080', withCredentials: true })

type PublicConfig = { authMode: 'phone' | 'microsoft'; pricePerGenerationPaise: number; currency: string; enabledModules: string[] }
type CurrentUser = { id: string; email: string | null; phone: string | null; displayName: string; isAdmin: boolean }
type UploadIntent = { assetId: string; uploadUrl: string }
type PaymentOrder = { orderId: string; razorpayOrderId: string; razorpayKeyId: string; amountPaise: number; currency: string }
type JobResult = { jobId: string; status: string }

type HistoryItem = {
	id: string
	kind: 'pixart' | 'comic' | 'payment'
	title: string
	detail: string
	createdAt: string
}

function unwrap<T>(response: { data: { data: T } }) { return response.data.data }
function money(paise: number, currency = 'INR') { return currency === 'INR' ? `₹${(paise / 100).toFixed(0)}` : `${currency} ${paise / 100}` }
function usePublicConfig() { return useQuery({ queryKey: ['public-config'], queryFn: async () => unwrap<PublicConfig>(await api.get('/api/v1/config/public')) }) }
function useCurrentUser() { return useQuery({ queryKey: ['me'], queryFn: async () => { try { return unwrap<CurrentUser>(await api.get('/api/v1/auth/me')) } catch (error) { if (axios.isAxiosError(error) && error.response?.status === 401) return null; throw error } } }) }

function loadHistory(): HistoryItem[] {
	try { return JSON.parse(localStorage.getItem('ai-day-history') || '[]') as HistoryItem[] } catch { return [] }
}
function saveHistory(items: HistoryItem[]) { localStorage.setItem('ai-day-history', JSON.stringify(items)) }

function Shell({ children, user, authMode, onLogout }: { children: ReactNode; user: CurrentUser | null; authMode: PublicConfig['authMode']; onLogout: () => void }) {
	return <main className="min-h-screen bg-[radial-gradient(circle_at_top,_rgba(34,211,238,0.18),_transparent_35%),linear-gradient(180deg,#020617_0%,#0f172a_45%,#020617_100%)] text-slate-100"><div className="mx-auto max-w-7xl px-6 py-5"><nav className="flex items-center justify-between rounded-3xl border border-white/10 bg-white/5 px-5 py-4 backdrop-blur-xl"><Link to="/" className="text-lg font-semibold tracking-tight">AI Day</Link><div className="flex items-center gap-4 text-sm text-slate-300"><Link to="/pixart" className="hover:text-white">PixArt</Link><Link to="/comic" className="hover:text-white">Comic</Link><Link to="/generations" className="hover:text-white">My creations</Link><Link to="/payments" className="hover:text-white">Payments</Link><span className="rounded-full border border-white/10 bg-white/5 px-3 py-1 text-xs uppercase tracking-[0.2em]">{authMode === 'phone' ? 'SMS auth' : 'Email auth'}</span>{user ? <button onClick={onLogout} className="rounded-full bg-cyan-300 px-4 py-2 font-semibold text-slate-950">Logout</button> : <Link to="/login" className="rounded-full bg-cyan-300 px-4 py-2 font-semibold text-slate-950">Login</Link>}</div></nav></div><div className="mx-auto max-w-7xl px-6 pb-12">{children}</div></main>
}

function HeroCard({ title, description, action, href }: { title: string; description: string; action: string; href: string }) {
	return <Link to={href} className="group rounded-[2rem] border border-white/10 bg-white/5 p-8 shadow-2xl shadow-cyan-950/20 backdrop-blur-xl transition hover:-translate-y-1 hover:border-cyan-300/40 hover:bg-white/10"><div className="flex h-full flex-col justify-between gap-8"><div><h2 className="text-2xl font-semibold">{title}</h2><p className="mt-3 max-w-md text-slate-300">{description}</p></div><div className="inline-flex w-fit rounded-full bg-cyan-300 px-5 py-3 font-semibold text-slate-950">{action}</div></div></Link>
}

function Landing({ config, user, onLogout }: { config: PublicConfig; user: CurrentUser | null; onLogout: () => void }) {
	return <Shell user={user} authMode={config.authMode} onLogout={onLogout}><section className="grid gap-8 py-12 lg:grid-cols-[1.15fr_0.85fr]"><div className="rounded-[2rem] border border-white/10 bg-white/5 p-10 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.25em] text-cyan-300">End-to-end test harness</p><h1 className="mt-5 max-w-3xl text-5xl font-bold tracking-tight sm:text-7xl">Exercise the backend flows from the browser.</h1><p className="mt-6 max-w-2xl text-lg text-slate-300">Switch auth modes in <span className="font-semibold text-white">.env</span>, sign in with SMS or Email, create a source upload, and send PixArt or comic jobs to the worker.</p><div className="mt-8 flex flex-wrap gap-3 text-sm text-slate-300"><span className="rounded-full border border-white/10 bg-white/5 px-4 py-2">Price: {money(config.pricePerGenerationPaise, config.currency)}</span><span className="rounded-full border border-white/10 bg-white/5 px-4 py-2">Modules: {config.enabledModules.join(', ')}</span><span className="rounded-full border border-white/10 bg-white/5 px-4 py-2">Mode: {config.authMode}</span></div><div className="mt-10 flex flex-wrap gap-4"><Link to="/login" className="rounded-full bg-cyan-300 px-5 py-3 font-semibold text-slate-950">Start sign in</Link><Link to="/generations" className="rounded-full border border-white/10 bg-white/5 px-5 py-3 font-semibold text-white">View activity</Link></div></div><div className="grid gap-5"><div className="rounded-[2rem] border border-white/10 bg-slate-950/60 p-6 backdrop-blur-xl"><p className="text-sm uppercase tracking-[0.2em] text-slate-400">Current session</p><h2 className="mt-3 text-2xl font-semibold">{user ? user.displayName : 'Not signed in'}</h2><p className="mt-2 text-slate-300">{user ? (user.email || user.phone || user.id) : 'Use the login flow to confirm the auth mode and session cookies.'}</p></div><HeroCard title="PixArt generator" description="Upload a selfie, reserve a credit, and queue a stylized portrait job." action="Open PixArt" href="/pixart" /><HeroCard title="Personal comic" description="Submit your biggest high and low, then let the worker process the comic job." action="Open comic" href="/comic" /></div></section><DashboardSummary config={config} /></Shell>
}

function DashboardSummary({ config }: { config: PublicConfig }) {
	const history = loadHistory()
	return <section className="grid gap-5 lg:grid-cols-3"><StatTile label="Auth mode" value={config.authMode === 'phone' ? 'SMS OTP' : 'Email login'} hint="Controlled by backend AUTH_MODE" /><StatTile label="Recent actions" value={String(history.length)} hint="Stored locally in the browser" /><StatTile label="Build target" value={money(config.pricePerGenerationPaise, config.currency)} hint="Price per generation" /></section>
}

function StatTile({ label, value, hint }: { label: string; value: string; hint: string }) { return <article className="rounded-[1.75rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl"><p className="text-sm uppercase tracking-[0.2em] text-slate-400">{label}</p><p className="mt-4 text-3xl font-semibold">{value}</p><p className="mt-2 text-sm text-slate-300">{hint}</p></article> }

function Login({ config, onLoggedIn }: { config: PublicConfig; onLoggedIn: () => void }) {
	const location = useLocation()
	const navigate = useNavigate()
	const queryClient = useQueryClient()
	const next = new URLSearchParams(location.search).get('next') || '/'
	const [phone, setPhone] = useState('+91')
	const [code, setCode] = useState('')
	const [email, setEmail] = useState('')
	const [sent, setSent] = useState(false)
	const [status, setStatus] = useState<string | null>(null)

	const sendOtp = useMutation({ mutationFn: async () => unwrap<{ status: string }>(await api.post('/api/v1/auth/phone/send-otp', { phone })) })
	const verifyOtp = useMutation({ mutationFn: async () => unwrap<{ userId: string }>(await api.post('/api/v1/auth/phone/verify-otp', { phone, code })) })

	useEffect(() => {
		const sessionToken = new URLSearchParams(location.search).get('session')
		if (sessionToken) {
			navigate(location.pathname, { replace: true })
		}
	}, [location.pathname, location.search, navigate])

	function continueWithEmail() {
		const search = new URLSearchParams({ next })
		if (email.trim()) search.set('email', email.trim())
		window.location.href = `${api.defaults.baseURL}/api/v1/auth/microsoft/start?${search.toString()}`
	}

	return <Shell user={null} authMode={config.authMode} onLogout={onLoggedIn}><section className="mx-auto max-w-2xl py-10"><div className="rounded-[2rem] border border-white/10 bg-white/5 p-8 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.25em] text-cyan-300">Sign in</p><h1 className="mt-4 text-4xl font-bold">{config.authMode === 'phone' ? 'Use SMS to continue' : 'Use email to continue'}</h1><p className="mt-3 max-w-xl text-slate-300">This page mirrors the backend auth mode from <span className="font-semibold text-white">/api/v1/config/public</span>. Switch <span className="font-semibold text-white">AUTH_MODE</span> in the env file to change the UI.</p>{config.authMode === 'phone' ? <form className="mt-8 space-y-4" onSubmit={async (event: FormEvent) => { event.preventDefault(); setStatus(null); if (!sent) { const response = await sendOtp.mutateAsync(); setSent(true); setStatus(response.status === 'otp_sent' ? 'OTP sent. Check the server logs in local development.' : 'OTP requested.'); return } await verifyOtp.mutateAsync(); await queryClient.invalidateQueries({ queryKey: ['me'] }); onLoggedIn(); navigate(next, { replace: true }) }}><label className="block text-sm font-medium text-slate-200">Phone number<input value={phone} onChange={(event) => setPhone(event.target.value)} placeholder="+919876543210" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none ring-0 placeholder:text-slate-500 focus:border-cyan-300" /></label>{sent && <label className="block text-sm font-medium text-slate-200">OTP code<input value={code} onChange={(event) => setCode(event.target.value)} placeholder="123456" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label>}<button type="submit" disabled={sendOtp.isPending || verifyOtp.isPending} className="w-full rounded-2xl bg-cyan-300 px-4 py-3 font-semibold text-slate-950 disabled:opacity-60">{sent ? 'Verify code' : 'Send OTP'}</button>{status && <p className="text-sm text-slate-300">{status}</p>}</form> : <div className="mt-8 space-y-4"><label className="block text-sm font-medium text-slate-200">Email address<input value={email} onChange={(event) => setEmail(event.target.value)} placeholder="user@example.com" className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><button type="button" onClick={continueWithEmail} className="w-full rounded-2xl bg-cyan-300 px-4 py-3 font-semibold text-slate-950">Continue with Email</button><p className="text-sm text-slate-300">This launches the backend Microsoft-style flow, which is mocked locally and returns to the app with a session cookie.</p></div>}</div></section></Shell>
}

function RequireAuth({ user, next, children }: { user: CurrentUser | null; next: string; children: ReactNode }) {
	if (!user) return <Navigate to={`/login?next=${encodeURIComponent(next)}`} replace />
	return <>{children}</>
}

function UploadAndGenerate({ module, user, config }: { module: 'pixel_portrait' | 'comic'; user: CurrentUser; config: PublicConfig }) {
	const queryClient = useQueryClient()
	const [file, setFile] = useState<File | null>(null)
	const [biggestHigh, setBiggestHigh] = useState('')
	const [biggestLow, setBiggestLow] = useState('')
	const [protagonistName, setProtagonistName] = useState(user.displayName || '')
	const [language, setLanguage] = useState('en')
	const [tone, setTone] = useState('warm')
	const [message, setMessage] = useState<string | null>(null)
	const [result, setResult] = useState<{ assetId: string; jobId: string } | null>(null)

	const uploadIntent = useMutation({ mutationFn: async () => unwrap<UploadIntent>(await api.post('/api/v1/assets/upload-url', { assetType: 'source_selfie', filename: file?.name || 'selfie.jpg', mimeType: file?.type || 'image/jpeg', sizeBytes: file?.size || 1 })) })
	const createGeneration = useMutation({ mutationFn: async (payload: { assetId: string; idempotencyKey: string }) => unwrap<JobResult>(await api.post(`/api/v1/generations/${module === 'pixel_portrait' ? 'pixart' : 'comic'}`, module === 'comic' ? { sourceAssetId: payload.assetId, idempotencyKey: payload.idempotencyKey, biggestHigh, biggestLow, protagonistName, language, tone } : { sourceAssetId: payload.assetId, idempotencyKey: payload.idempotencyKey })) })

	async function uploadFile(intent: UploadIntent) {
		if (!file) return
		if (intent.uploadUrl.startsWith('http://') || intent.uploadUrl.startsWith('https://')) {
			await fetch(intent.uploadUrl, { method: 'PUT', headers: { 'Content-Type': file.type || 'application/octet-stream' }, body: file })
		}
	}

	function addHistory(entry: HistoryItem) {
		const nextHistory = [entry, ...loadHistory()].slice(0, 20)
		saveHistory(nextHistory)
		window.dispatchEvent(new Event('storage'))
	}

	return <section className="grid gap-6 py-10 lg:grid-cols-[1.1fr_0.9fr]"><div className="rounded-[2rem] border border-white/10 bg-white/5 p-8 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.25em] text-cyan-300">{module === 'comic' ? 'Comic job' : 'PixArt job'}</p><h1 className="mt-4 text-4xl font-bold">{module === 'comic' ? 'Turn lived moments into a comic' : 'Turn a selfie into stylized art'}</h1><p className="mt-3 text-slate-300">This flow exercises upload intent creation, authenticated generation submission, and worker processing.</p><form className="mt-8 space-y-5" onSubmit={async (event) => { event.preventDefault(); if (!file) { setMessage('Choose an image first.'); return } setMessage(null); const intent = await uploadIntent.mutateAsync(); await uploadFile(intent); const idempotencyKey = crypto.randomUUID(); const job = await createGeneration.mutateAsync({ assetId: intent.assetId, idempotencyKey }); setResult({ assetId: intent.assetId, jobId: job.jobId }); addHistory({ id: job.jobId, kind: module === 'comic' ? 'comic' : 'pixart', title: module === 'comic' ? 'Comic submitted' : 'PixArt submitted', detail: `Asset ${intent.assetId} queued as ${job.status}.`, createdAt: new Date().toISOString() }); setMessage(`Generation queued. Check the worker and admin pages for status.`); await queryClient.invalidateQueries({ queryKey: ['me'] }) }}><label className="block text-sm font-medium text-slate-200">Source selfie<input type="file" accept="image/jpeg,image/png,image/webp" onChange={(event) => setFile(event.target.files?.[0] || null)} className="mt-2 block w-full rounded-2xl border border-dashed border-white/10 bg-slate-950/70 px-4 py-3 text-sm text-slate-300 file:mr-4 file:rounded-full file:border-0 file:bg-cyan-300 file:px-4 file:py-2 file:font-semibold file:text-slate-950" /></label>{module === 'comic' && <div className="grid gap-4 md:grid-cols-2"><label className="block text-sm font-medium text-slate-200">Biggest high<textarea value={biggestHigh} onChange={(event) => setBiggestHigh(event.target.value)} className="mt-2 min-h-28 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><label className="block text-sm font-medium text-slate-200">Biggest low<textarea value={biggestLow} onChange={(event) => setBiggestLow(event.target.value)} className="mt-2 min-h-28 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><label className="block text-sm font-medium text-slate-200">Protagonist name<input value={protagonistName} onChange={(event) => setProtagonistName(event.target.value)} className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><label className="block text-sm font-medium text-slate-200">Language<input value={language} onChange={(event) => setLanguage(event.target.value)} className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label><label className="block text-sm font-medium text-slate-200">Tone<input value={tone} onChange={(event) => setTone(event.target.value)} className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3 outline-none placeholder:text-slate-500 focus:border-cyan-300" /></label></div>}<button type="submit" disabled={uploadIntent.isPending || createGeneration.isPending} className="rounded-2xl bg-cyan-300 px-5 py-3 font-semibold text-slate-950 disabled:opacity-60">Queue generation</button>{message && <p className="text-sm text-slate-300">{message}</p>}</form></div><div className="space-y-5"><Panel title="What the backend will do" subtitle="The worker polls queued jobs and updates job state in PostgreSQL." lines={[`Module: ${module}`, `Storage provider: ${config.authMode === 'phone' ? 'local browser session + backend storage' : 'backend configured'}`, `Credits: reserved on generation creation`, `Final status: available in admin generation view`]} /><Panel title="Latest submission" subtitle="A local history entry helps you verify the flow without a list endpoint." lines={result ? [`Asset ID: ${result.assetId}`, `Job ID: ${result.jobId}`] : ['No jobs submitted yet.']} /><HistoryPanel /></div></section>
}

function Panel({ title, subtitle, lines }: { title: string; subtitle: string; lines: string[] }) { return <article className="rounded-[1.75rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-400">{title}</p><p className="mt-2 text-sm text-slate-300">{subtitle}</p><ul className="mt-4 space-y-2 text-sm text-slate-200">{lines.map((line) => <li key={line} className="rounded-xl border border-white/5 bg-slate-950/50 px-3 py-2">{line}</li>)}</ul></article> }

function HistoryPanel() {
	const [history, setHistory] = useState<HistoryItem[]>(loadHistory())
	useEffect(() => { const sync = () => setHistory(loadHistory()); window.addEventListener('storage', sync); return () => window.removeEventListener('storage', sync) }, [])
	return <article className="rounded-[1.75rem] border border-white/10 bg-white/5 p-6 backdrop-blur-xl"><p className="text-sm font-semibold uppercase tracking-[0.2em] text-slate-400">Recent activity</p><div className="mt-4 space-y-3">{history.length ? history.map((item) => <div key={item.id} className="rounded-xl border border-white/5 bg-slate-950/50 px-4 py-3"><div className="flex items-center justify-between gap-4"><div><p className="font-semibold text-white">{item.title}</p><p className="text-sm text-slate-300">{item.detail}</p></div><time className="text-xs text-slate-400">{new Date(item.createdAt).toLocaleString()}</time></div></div>) : <p className="text-sm text-slate-300">No local activity yet.</p>}</div></article>
}

function PaymentsPage({ user, config }: { user: CurrentUser; config: PublicConfig }) {
	const [module, setModule] = useState<'pixel_portrait' | 'comic'>('pixel_portrait')
	const [order, setOrder] = useState<PaymentOrder | null>(null)
	const mutation = useMutation({ mutationFn: async () => unwrap<PaymentOrder>(await api.post('/api/v1/payments/orders', { module, quantity: 1 })) })
	return <RequireAuth user={user} next="/payments"><section className="grid gap-6 py-10 lg:grid-cols-[0.9fr_1.1fr]"><div className="rounded-[2rem] border border-white/10 bg-white/5 p-8 backdrop-blur-xl"><h1 className="text-4xl font-bold">Payments</h1><p className="mt-3 text-slate-300">Create a mock Razorpay order to validate the payment endpoint and continue testing the generation flow.</p><div className="mt-6 space-y-4"><label className="block text-sm font-medium text-slate-200">Module<select value={module} onChange={(event) => setModule(event.target.value as 'pixel_portrait' | 'comic')} className="mt-2 w-full rounded-2xl border border-white/10 bg-slate-950/70 px-4 py-3"><option value="pixel_portrait">PixArt</option><option value="comic">Comic</option></select></label><button onClick={async () => { const nextOrder = await mutation.mutateAsync(); setOrder(nextOrder); const history = [{ id: nextOrder.orderId, kind: 'payment' as const, title: 'Payment order created', detail: `${nextOrder.razorpayOrderId} for ${money(nextOrder.amountPaise, nextOrder.currency)}`, createdAt: new Date().toISOString() }, ...loadHistory()].slice(0, 20); saveHistory(history); window.dispatchEvent(new Event('storage')) }} className="rounded-2xl bg-cyan-300 px-5 py-3 font-semibold text-slate-950">Create payment order</button></div></div><div className="space-y-5"><Panel title="Configured price" subtitle="Matches the backend public config." lines={[money(config.pricePerGenerationPaise, config.currency)]} /><Panel title="Latest order" subtitle="Use this payload to confirm the payment route responds." lines={order ? [`Order ID: ${order.orderId}`, `Razorpay order: ${order.razorpayOrderId}`, `Key ID: ${order.razorpayKeyId}`, `Amount: ${money(order.amountPaise, order.currency)}`] : ['No order created yet.']} /></div></section></RequireAuth>
}

function GenerationsPage({ user }: { user: CurrentUser }) {
	return <RequireAuth user={user} next="/generations"><section className="py-10"><div className="rounded-[2rem] border border-white/10 bg-white/5 p-8 backdrop-blur-xl"><h1 className="text-4xl font-bold">My creations</h1><p className="mt-3 text-slate-300">The backend currently exposes creation endpoints rather than per-user listing, so this page surfaces local test history and links back to the creator pages.</p><div className="mt-6 flex flex-wrap gap-3"><Link to="/pixart" className="rounded-full bg-cyan-300 px-5 py-3 font-semibold text-slate-950">New PixArt</Link><Link to="/comic" className="rounded-full border border-white/10 bg-white/5 px-5 py-3 font-semibold text-white">New comic</Link></div><div className="mt-8"><HistoryPanel /></div></div></section></RequireAuth>
}

function LoginRoute({ config, onLoggedIn }: { config: PublicConfig; onLoggedIn: () => void }) { return <Login config={config} onLoggedIn={onLoggedIn} /> }

export default function App() {
	const configQuery = usePublicConfig()
	const userQuery = useCurrentUser()
	const queryClient = useQueryClient()
	const navigate = useNavigate()

	const config = configQuery.data
	const user = userQuery.data ?? null

	const signOut = async () => {
		await api.post('/api/v1/auth/logout')
		await queryClient.invalidateQueries({ queryKey: ['me'] })
		navigate('/login', { replace: true })
	}

	const loading = configQuery.isLoading || configQuery.isFetching
	if (loading || !config) {
		return <main className="min-h-screen bg-slate-950 text-slate-100"><div className="mx-auto flex min-h-screen max-w-4xl items-center justify-center px-6 text-slate-300">Loading AI Day…</div></main>
	}

	return <Routes><Route path="/" element={<Landing config={config} user={user} onLogout={signOut} />} /><Route path="/login" element={<LoginRoute config={config} onLoggedIn={() => void 0} />} /><Route path="/pixart" element={<RequireAuth user={user} next="/pixart"><Shell user={user} authMode={config.authMode} onLogout={signOut}><UploadAndGenerate module="pixel_portrait" user={user!} config={config} /></Shell></RequireAuth>} /><Route path="/comic" element={<RequireAuth user={user} next="/comic"><Shell user={user} authMode={config.authMode} onLogout={signOut}><UploadAndGenerate module="comic" user={user!} config={config} /></Shell></RequireAuth>} /><Route path="/generations" element={<GenerationsPage user={user!} />} /><Route path="/payments" element={<PaymentsPage user={user!} config={config} />} /><Route path="*" element={<Navigate to="/" replace />} /></Routes>
}
