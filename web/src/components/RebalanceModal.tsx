import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
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
        aria-describedby="rebal-desc"
      >
        <DialogHeader>
          <DialogTitle id="rebal-title">Rebalance</DialogTitle>
          <DialogDescription id="rebal-desc">
            Two-step flow: preview estimates, then confirm. This build may not submit
            exchange orders for rebalance.
          </DialogDescription>
        </DialogHeader>
        {!preview ? (
          <Button type="button" onClick={runPreview} disabled={busy}>
            {busy ? 'Loading…' : 'Run preview'}
          </Button>
        ) : (
          <div className="space-y-3 rounded-md border border-border bg-muted/40 p-3 text-sm">
            <p>
              Suggested legs: {preview.suggested_adjustments.length} (empty in v0).
            </p>
            {preview.disclaimer ? (
              <p className="text-xs text-amber-200/90">{preview.disclaimer}</p>
            ) : null}
            <DialogFooter className="gap-2 sm:justify-start">
              <Button
                type="button"
                className="bg-amber-600 text-white hover:bg-amber-500"
                onClick={confirm}
                disabled={busy}
              >
                Confirm
              </Button>
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
            </DialogFooter>
          </div>
        )}
        {msg ? <p className="text-sm text-primary">{msg}</p> : null}
        {!preview ? (
          <Button type="button" variant="ghost" className="px-0" onClick={onClose}>
            Close
          </Button>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
