import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-lg"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
        aria-describedby="close-desc"
      >
        <DialogHeader>
          <DialogTitle id="close-title">Close position</DialogTitle>
          <p id="close-desc" className="text-sm text-muted-foreground">
            {label}
          </p>
        </DialogHeader>
        <p className="text-xs text-amber-200/90">
          Estimates only. Live reduce-only market orders can slip; verify size and venue
          state before confirming.
        </p>
        {!preview ? (
          <Button type="button" onClick={runPreview} disabled={busy}>
            {busy ? 'Loading…' : 'Preview close'}
          </Button>
        ) : (
          <div className="space-y-2 rounded-md border border-border bg-muted/40 p-3 text-sm">
            <p>
              Est. exit P/L (floating){' '}
              <span className="font-mono">{preview.estimated_exit_pnl}</span>
            </p>
            <p className="text-xs text-muted-foreground">{preview.wait_vs_close_guidance}</p>
            {preview.labels?.length ? (
              <ul className="list-inside list-disc text-xs text-amber-200/80">
                {preview.labels.map((l) => (
                  <li key={l}>{l}</li>
                ))}
              </ul>
            ) : null}
            <DialogFooter className="gap-2 sm:justify-start">
              <Button type="button" variant="destructive" onClick={confirm} disabled={busy}>
                Confirm close
              </Button>
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
            </DialogFooter>
          </div>
        )}
        {err ? <p className="text-sm text-destructive">{err}</p> : null}
        {!preview ? (
          <Button type="button" variant="ghost" className="px-0" onClick={onClose}>
            Close
          </Button>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
