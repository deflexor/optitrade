import { useState } from 'react'
import { api } from '../api/client'

type Preview = {
  preview_token: string
  suggested_adjustments: unknown[]
  disclaimer?: string
}

export default function RebalanceModal({ onClose }: { onClose: () => void }) {
  const [preview, setPreview] = useState<Preview | null>(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)

  async function runPreview() {
    setBusy(true)
    setMsg(null)
    try {
      const { data } = await api.post<Preview>('/rebalance/preview', {})
      setPreview(data)
    } catch {
      setMsg('Preview failed.')
    } finally {
      setBusy(false)
    }
  }

  async function confirm() {
    if (!preview?.preview_token) return
    setBusy(true)
    try {
      await api.post('/rebalance/confirm', {
        preview_token: preview.preview_token,
      })
      setMsg('Acknowledged. No rebalance orders were sent in this build.')
    } catch {
      setMsg('Confirm failed — preview may have expired.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="rebal-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-xl">
        <h2 id="rebal-title" className="text-lg font-semibold text-slate-100">
          Rebalance
        </h2>
        <p className="mt-2 text-sm text-slate-400">
          Two-step flow: preview estimates, then confirm. This build may not
          submit exchange orders for rebalance.
        </p>
        {!preview ? (
          <button
            type="button"
            className="mt-4 rounded-md bg-slate-700 px-4 py-2 text-sm text-white hover:bg-slate-600 disabled:opacity-50"
            onClick={runPreview}
            disabled={busy}
          >
            {busy ? 'Loading…' : 'Run preview'}
          </button>
        ) : (
          <div className="mt-4 space-y-3 rounded-md border border-slate-800 bg-slate-950/60 p-3 text-sm text-slate-200">
            <p>
              Suggested legs: {preview.suggested_adjustments.length} (empty in
              v0).
            </p>
            {preview.disclaimer ? (
              <p className="text-xs text-amber-200/90">{preview.disclaimer}</p>
            ) : null}
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="rounded-md bg-amber-600 px-4 py-2 text-sm text-white hover:bg-amber-500 disabled:opacity-50"
                onClick={confirm}
                disabled={busy}
              >
                Confirm
              </button>
              <button
                type="button"
                className="rounded-md border border-slate-600 px-4 py-2 text-sm text-slate-200 hover:bg-slate-800"
                onClick={onClose}
              >
                Cancel
              </button>
            </div>
          </div>
        )}
        {msg ? <p className="mt-3 text-sm text-emerald-200/90">{msg}</p> : null}
        {!preview ? (
          <button
            type="button"
            className="mt-4 text-sm text-slate-500 hover:text-slate-300"
            onClick={onClose}
          >
            Close
          </button>
        ) : null}
      </div>
    </div>
  )
}
