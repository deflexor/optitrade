import { useState } from 'react'
import { api } from '../api/client'

type Preview = {
  preview_token: string
  estimated_exit_pnl: string
  wait_vs_close_guidance: string
  assumptions?: string[]
  labels?: string[]
}

export default function CloseModal({
  positionId,
  label,
  onClose,
}: {
  positionId: string
  label: string
  onClose: () => void
}) {
  const [preview, setPreview] = useState<Preview | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  async function runPreview() {
    setBusy(true)
    setErr(null)
    try {
      const { data } = await api.post<Preview>(
        `/positions/${encodeURIComponent(positionId)}/close/preview`,
        {},
      )
      setPreview(data)
    } catch {
      setErr('Preview failed — position may be stale or exchange unavailable.')
    } finally {
      setBusy(false)
    }
  }

  async function confirm() {
    if (!preview?.preview_token) return
    setBusy(true)
    setErr(null)
    try {
      await api.post(
        `/positions/${encodeURIComponent(positionId)}/close/confirm`,
        { preview_token: preview.preview_token },
      )
      onClose()
    } catch {
      setErr('Confirm rejected — re-preview before trying again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 px-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="close-title"
    >
      <div className="max-h-[90vh] w-full max-w-lg overflow-y-auto rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-xl">
        <h2 id="close-title" className="text-lg font-semibold text-slate-100">
          Close position
        </h2>
        <p className="mt-1 text-sm text-slate-400">{label}</p>
        <p className="mt-3 text-xs text-amber-200/90">
          Estimates only. Live reduce-only market orders can slip; verify size
          and venue state before confirming.
        </p>
        {!preview ? (
          <button
            type="button"
            className="mt-4 rounded-md bg-slate-700 px-4 py-2 text-sm text-white hover:bg-slate-600 disabled:opacity-50"
            onClick={runPreview}
            disabled={busy}
          >
            {busy ? 'Loading…' : 'Preview close'}
          </button>
        ) : (
          <div className="mt-4 space-y-2 rounded-md border border-slate-800 bg-slate-950/60 p-3 text-sm text-slate-200">
            <p>
              Est. exit P/L (floating){' '}
              <span className="font-mono">{preview.estimated_exit_pnl}</span>
            </p>
            <p className="text-xs text-slate-400">
              {preview.wait_vs_close_guidance}
            </p>
            {preview.labels?.length ? (
              <ul className="list-inside list-disc text-xs text-amber-200/80">
                {preview.labels.map((l) => (
                  <li key={l}>{l}</li>
                ))}
              </ul>
            ) : null}
            <div className="flex flex-wrap gap-2 pt-2">
              <button
                type="button"
                className="rounded-md bg-rose-600 px-4 py-2 text-sm text-white hover:bg-rose-500 disabled:opacity-50"
                onClick={confirm}
                disabled={busy}
              >
                Confirm close
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
        {err ? <p className="mt-3 text-sm text-rose-300/90">{err}</p> : null}
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
