import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { api } from '../api/client'
import { Link } from 'react-router-dom'

type FieldDef = {
  id: string
  label: string
  help: string
  kind: string
  required: boolean
  options?: { value: string; label: string }[]
}

type SecretView = { masked: string; configured: boolean }

type SettingsResp = {
  fields: FieldDef[]
  values: Record<string, unknown>
  warnings: string[]
}

export default function SettingsPage() {
  const [payload, setPayload] = useState<SettingsResp | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<{
    provider: string
    deribit_use_mainnet: boolean
    okx_demo: boolean
    currencies: string
    deribit_client_id: string
    deribit_client_secret: string
    okx_api_key: string
    okx_secret_key: string
    okx_passphrase: string
  } | null>(null)

  const load = useCallback(async () => {
    setErr(null)
    try {
      const { data } = await api.get<SettingsResp>('/settings')
      setPayload(data)
      const v = data.values
      const sec = (k: string) =>
        (v[k] as SecretView | undefined)?.configured
          ? ''
          : ((v[k] as SecretView | undefined)?.masked ?? '')
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        deribit_client_id: sec('deribit_client_id'),
        deribit_client_secret: sec('deribit_client_secret'),
        okx_api_key: sec('okx_api_key'),
        okx_secret_key: sec('okx_secret_key'),
        okx_passphrase: sec('okx_passphrase'),
      })
    } catch {
      setErr('Could not load settings.')
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!form) return
    setSaving(true)
    setErr(null)
    try {
      const secrets: Record<string, string> = {}
      if (form.deribit_client_id.trim())
        secrets.deribit_client_id = form.deribit_client_id.trim()
      if (form.deribit_client_secret.trim())
        secrets.deribit_client_secret = form.deribit_client_secret.trim()
      if (form.okx_api_key.trim()) secrets.okx_api_key = form.okx_api_key.trim()
      if (form.okx_secret_key.trim()) secrets.okx_secret_key = form.okx_secret_key.trim()
      if (form.okx_passphrase.trim()) secrets.okx_passphrase = form.okx_passphrase.trim()

      const { data } = await api.put<SettingsResp>('/settings', {
        provider: form.provider,
        deribit_use_mainnet: form.deribit_use_mainnet,
        okx_demo: form.okx_demo,
        currencies: form.currencies.trim(),
        secrets,
      })
      setPayload(data)
      const v = data.values
      const sec = (k: string) =>
        (v[k] as SecretView | undefined)?.configured
          ? ''
          : ((v[k] as SecretView | undefined)?.masked ?? '')
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        deribit_client_id: sec('deribit_client_id'),
        deribit_client_secret: sec('deribit_client_secret'),
        okx_api_key: sec('okx_api_key'),
        okx_secret_key: sec('okx_secret_key'),
        okx_passphrase: sec('okx_passphrase'),
      })
    } catch (e: unknown) {
      const ax = e as { response?: { data?: { message?: string } } }
      setErr(ax.response?.data?.message ?? 'Save failed.')
    } finally {
      setSaving(false)
    }
  }

  if (!form || !payload) {
    return (
      <div className="space-y-4">
        <Link to="/" className="text-sm text-emerald-400 hover:text-emerald-300">
          ← Overview
        </Link>
        <p className="text-slate-500">{err ?? 'Loading…'}</p>
      </div>
    )
  }

  const prov = form.provider

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-slate-100">Settings</h2>
        <Link to="/" className="text-sm text-emerald-400 hover:text-emerald-300">
          ← Overview
        </Link>
      </div>

      {payload.warnings.length > 0 ? (
        <div
          className="space-y-2 rounded-lg border border-amber-500/40 bg-amber-950/30 p-4 text-sm text-amber-100"
          role="alert"
        >
          {payload.warnings.map((w) => (
            <p key={w}>{w}</p>
          ))}
        </div>
      ) : null}

      {err ? (
        <p className="text-sm text-rose-300" role="alert">
          {err}
        </p>
      ) : null}

      <form onSubmit={(e) => void onSubmit(e)} className="max-w-xl space-y-6">
        <div>
          <label className="block text-sm font-medium text-slate-300">Exchange</label>
          <select
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-900 px-3 py-2 text-slate-100"
            value={prov}
            onChange={(e) => setForm({ ...form, provider: e.target.value })}
          >
            <option value="deribit">Deribit</option>
            <option value="okx">OKX</option>
          </select>
          <p className="mt-1 text-xs text-slate-500">Venue for this operator account.</p>
        </div>

        {prov === 'deribit' ? (
          <div className="space-y-4 rounded-lg border border-slate-800 bg-slate-900/40 p-4">
            <label className="flex items-center gap-2 text-sm text-slate-200">
              <input
                type="checkbox"
                checked={form.deribit_use_mainnet}
                onChange={(e) =>
                  setForm({ ...form, deribit_use_mainnet: e.target.checked })
                }
              />
              Use Deribit mainnet (off = testnet)
            </label>
            <div>
              <label className="block text-sm text-slate-400">Client ID</label>
              <input
                type="password"
                autoComplete="off"
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100"
                value={form.deribit_client_id}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, deribit_client_id: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400">Client secret</label>
              <input
                type="password"
                autoComplete="off"
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm text-slate-100"
                value={form.deribit_client_secret}
                placeholder="Leave blank to keep saved value"
                onChange={(e) =>
                  setForm({ ...form, deribit_client_secret: e.target.value })
                }
              />
            </div>
          </div>
        ) : (
          <div className="space-y-4 rounded-lg border border-slate-800 bg-slate-900/40 p-4">
            <label className="flex items-center gap-2 text-sm text-slate-200">
              <input
                type="checkbox"
                checked={form.okx_demo}
                onChange={(e) => setForm({ ...form, okx_demo: e.target.checked })}
              />
              OKX demo trading (Demo Trading API keys + simulated header)
            </label>
            <div>
              <label className="block text-sm text-slate-400">API key</label>
              <input
                type="password"
                autoComplete="off"
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm"
                value={form.okx_api_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_api_key: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400">Secret key</label>
              <input
                type="password"
                autoComplete="off"
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm"
                value={form.okx_secret_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_secret_key: e.target.value })}
              />
            </div>
            <div>
              <label className="block text-sm text-slate-400">Passphrase</label>
              <input
                type="password"
                autoComplete="off"
                className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 font-mono text-sm"
                value={form.okx_passphrase}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_passphrase: e.target.value })}
              />
            </div>
          </div>
        )}

        <div>
          <label className="block text-sm text-slate-400">Overview currencies</label>
          <input
            className="mt-1 w-full rounded-md border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-100"
            value={form.currencies}
            placeholder="e.g. BTC, ETH"
            onChange={(e) => setForm({ ...form, currencies: e.target.value })}
          />
          <p className="mt-1 text-xs text-slate-500">
            Comma-separated; blank uses server default (env OPTITRADE_DASHBOARD_CURRENCIES or BTC, ETH).
          </p>
        </div>

        <button
          type="submit"
          disabled={saving}
          className="rounded-md bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-500 disabled:opacity-50"
        >
          {saving ? 'Saving…' : 'Save settings'}
        </button>
      </form>
    </div>
  )
}
