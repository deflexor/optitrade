import { useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

export default function Login() {
  const { username, login, loginError, clearError } = useAuthStore()
  const loc = useLocation()
  const [u, setU] = useState('')
  const [p, setP] = useState('')
  const [busy, setBusy] = useState(false)

  if (username) {
    const to = (loc.state as { from?: string })?.from ?? '/'
    return <Navigate to={to} replace />
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    clearError()
    setBusy(true)
    try {
      await login(u.trim(), p)
    } catch {
      /* error surfaced via store */
    } finally {
      setBusy(false)
      setP('')
    }
  }

  return (
    <div className="min-h-screen bg-slate-950 px-4 py-16 text-slate-100">
      <div className="mx-auto max-w-sm">
      <h1 className="mb-6 text-xl font-semibold tracking-tight text-slate-100">
        Operator sign-in
      </h1>
      <form onSubmit={onSubmit} className="space-y-4">
        <div>
          <label className="mb-1 block text-sm text-slate-400" htmlFor="user">
            Username
          </label>
          <input
            id="user"
            autoComplete="username"
            className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-slate-100 outline-none ring-emerald-500/40 focus:ring-2"
            value={u}
            onChange={(e) => setU(e.target.value)}
            required
          />
        </div>
        <div>
          <label className="mb-1 block text-sm text-slate-400" htmlFor="pass">
            Password
          </label>
          <input
            id="pass"
            type="password"
            autoComplete="current-password"
            className="w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-slate-100 outline-none ring-emerald-500/40 focus:ring-2"
            value={p}
            onChange={(e) => setP(e.target.value)}
            required
          />
        </div>
        {loginError ? (
          <p className="text-sm text-amber-400/90" role="alert">
            {loginError}
          </p>
        ) : null}
        <button
          type="submit"
          disabled={busy}
          className="w-full rounded-md bg-emerald-600 py-2.5 text-sm font-medium text-white shadow hover:bg-emerald-500 disabled:opacity-60"
        >
          {busy ? 'Signing in…' : 'Sign in'}
        </button>
      </form>
      </div>
    </div>
  )
}
