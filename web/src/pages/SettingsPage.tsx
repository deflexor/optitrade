import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { api } from '../api/client'

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
  /** Present as [] from server; null if older server or malformed JSON */
  warnings?: string[] | null
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
    max_loss_equity_pct: number
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
      const rawMax = v.max_loss_equity_pct
      let maxLoss = 10
      if (typeof rawMax === 'number' && Number.isFinite(rawMax)) {
        maxLoss = Math.round(rawMax)
      } else if (typeof rawMax === 'string' && rawMax.trim() !== '') {
        const p = parseInt(rawMax, 10)
        if (!Number.isNaN(p)) maxLoss = p
      }
      maxLoss = Math.min(50, Math.max(1, maxLoss))
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        max_loss_equity_pct: maxLoss,
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

      const maxLoss = Math.min(50, Math.max(1, Math.round(form.max_loss_equity_pct)))
      const { data } = await api.put<SettingsResp>('/settings', {
        provider: form.provider,
        deribit_use_mainnet: form.deribit_use_mainnet,
        okx_demo: form.okx_demo,
        currencies: form.currencies.trim(),
        max_loss_equity_pct: maxLoss,
        secrets,
      })
      setPayload(data)
      const v = data.values
      const sec = (k: string) =>
        (v[k] as SecretView | undefined)?.configured
          ? ''
          : ((v[k] as SecretView | undefined)?.masked ?? '')
      const rawMaxAfter = v.max_loss_equity_pct
      let maxLossAfter = 10
      if (typeof rawMaxAfter === 'number' && Number.isFinite(rawMaxAfter)) {
        maxLossAfter = Math.round(rawMaxAfter)
      } else if (typeof rawMaxAfter === 'string' && rawMaxAfter.trim() !== '') {
        const p = parseInt(rawMaxAfter, 10)
        if (!Number.isNaN(p)) maxLossAfter = p
      }
      maxLossAfter = Math.min(50, Math.max(1, maxLossAfter))
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        max_loss_equity_pct: maxLossAfter,
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
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
        <p className="text-muted-foreground">{err ?? 'Loading…'}</p>
      </div>
    )
  }

  const prov = form.provider

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-foreground">Settings</h2>
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
      </div>

      {(payload.warnings?.length ?? 0) > 0 ? (
        <Alert
          variant="default"
          className="border-amber-500/40 bg-amber-950/30 text-amber-100"
        >
          <AlertDescription className="space-y-2 text-amber-100">
            {(payload.warnings ?? []).map((w) => (
              <p key={w}>{w}</p>
            ))}
          </AlertDescription>
        </Alert>
      ) : null}

      {err ? (
        <p className="text-sm text-destructive" role="alert">
          {err}
        </p>
      ) : null}

      <form onSubmit={(e) => void onSubmit(e)} className="max-w-xl space-y-6">
        <div className="space-y-2">
          <Label>Exchange</Label>
          <Select
            value={prov}
            onValueChange={(value) => setForm({ ...form, provider: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="Exchange" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="deribit">Deribit</SelectItem>
              <SelectItem value="okx">OKX</SelectItem>
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">Venue for this operator account.</p>
        </div>

        <Separator />

        {prov === 'deribit' ? (
          <div className="space-y-4 rounded-lg border border-border bg-muted/60 p-4">
            <div className="flex items-center gap-2">
              <Switch
                id="deribit-mainnet"
                checked={form.deribit_use_mainnet}
                onCheckedChange={(checked) =>
                  setForm({ ...form, deribit_use_mainnet: checked })
                }
              />
              <Label htmlFor="deribit-mainnet" className="text-sm font-normal">
                Use Deribit mainnet (off = testnet)
              </Label>
            </div>
            <div className="space-y-2">
              <Label htmlFor="deribit-client-id">Client ID</Label>
              <Input
                id="deribit-client-id"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.deribit_client_id}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, deribit_client_id: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="deribit-client-secret">Client secret</Label>
              <Input
                id="deribit-client-secret"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.deribit_client_secret}
                placeholder="Leave blank to keep saved value"
                onChange={(e) =>
                  setForm({ ...form, deribit_client_secret: e.target.value })
                }
              />
            </div>
          </div>
        ) : (
          <div className="space-y-4 rounded-lg border border-border bg-muted/60 p-4">
            <div className="flex items-center gap-2">
              <Switch
                id="okx-demo"
                checked={form.okx_demo}
                onCheckedChange={(checked) => setForm({ ...form, okx_demo: checked })}
              />
              <Label htmlFor="okx-demo" className="text-sm font-normal">
                OKX demo trading (Demo Trading API keys + simulated header)
              </Label>
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-api-key">API key</Label>
              <Input
                id="okx-api-key"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_api_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_api_key: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-secret-key">Secret key</Label>
              <Input
                id="okx-secret-key"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_secret_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_secret_key: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-passphrase">Passphrase</Label>
              <Input
                id="okx-passphrase"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_passphrase}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_passphrase: e.target.value })}
              />
            </div>
          </div>
        )}

        <Separator />

        <div className="space-y-2">
          <Label htmlFor="currencies">Overview currencies</Label>
          <Input
            id="currencies"
            value={form.currencies}
            placeholder="e.g. BTC, ETH"
            onChange={(e) => setForm({ ...form, currencies: e.target.value })}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated; blank uses server default (env OPTITRADE_DASHBOARD_CURRENCIES or BTC,
            ETH).
          </p>
        </div>

        <div className="space-y-2">
          <Label htmlFor="max-loss-pct">Max loss % of equity</Label>
          <Input
            id="max-loss-pct"
            type="number"
            inputMode="numeric"
            min={1}
            max={50}
            className="max-w-[8rem] font-mono text-sm"
            value={form.max_loss_equity_pct}
            onChange={(e) => {
              const p = parseInt(e.target.value, 10)
              setForm({
                ...form,
                max_loss_equity_pct: Number.isNaN(p) ? form.max_loss_equity_pct : p,
              })
            }}
          />
          <p className="text-xs text-muted-foreground">
            Cap on max loss per opportunity vs account equity (1–50). Used for opportunity ranking
            and gates.
          </p>
        </div>

        <Button type="submit" disabled={saving}>
          {saving ? 'Saving…' : 'Save settings'}
        </Button>
      </form>
    </div>
  )
}
