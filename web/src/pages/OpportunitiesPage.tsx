import { useCallback, useEffect, useState } from 'react'
import { AlertCircle } from 'lucide-react'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  getOpportunities,
  postOpportunityCancel,
  postOpportunityClose,
  postOpportunityOpen,
  type OpportunitiesResponse,
  type OpportunityRow,
} from '@/api/trading'
import { parseApiError } from '@/api/parseApiError'

function statusLabel(s: string): string {
  return s || '—'
}

function rowActions(r: OpportunityRow): { open?: boolean; cancel?: boolean; close?: boolean } {
  const st = (r.status || '').toLowerCase()
  const rec = (r.recommendation || '').toLowerCase()
  return {
    open: st === 'candidate' && rec === 'open',
    cancel: st === 'opening',
    close: st === 'active' || st === 'partial',
  }
}

export default function OpportunitiesPage() {
  const [payload, setPayload] = useState<OpportunitiesResponse | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [busyId, setBusyId] = useState<string | null>(null)

  const applyData = useCallback((d: OpportunitiesResponse) => {
    setPayload({
      ...d,
      rows: d.rows ?? [],
    })
    setErr(null)
  }, [])

  const load = useCallback(async () => {
    try {
      const d = await getOpportunities()
      applyData(d)
    } catch (e) {
      const p = parseApiError(e)
      setErr(p?.message ?? 'Could not load opportunities.')
    }
  }, [applyData])

  useEffect(() => {
    let cancelled = false
    let poll: number | null = null
    const es = new EventSource('/api/v1/opportunities/stream')

    es.onmessage = (ev) => {
      try {
        const d = JSON.parse(ev.data) as OpportunitiesResponse
        if (!cancelled) {
          applyData(d)
        }
      } catch {
        /* ignore malformed chunk */
      }
    }

    es.onerror = () => {
      es.close()
      if (!cancelled) {
        void load()
        poll = window.setInterval(() => void load(), 2000)
      }
    }

    void load()

    return () => {
      cancelled = true
      es.close()
      if (poll !== null) {
        window.clearInterval(poll)
      }
    }
  }, [applyData, load])

  async function runAction(
    id: string,
    fn: (i: string) => Promise<void>,
  ) {
    setBusyId(id)
    try {
      await fn(id)
      await load()
    } catch (e) {
      const p = parseApiError(e)
      setErr(p?.message ?? 'Action failed.')
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-lg font-semibold text-foreground">Opportunities</h2>

      {err ? (
        <p className="text-sm text-amber-300/90" role="alert">
          {err}
        </p>
      ) : null}

      {!payload ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : payload.disabled ? (
        <Alert variant="destructive">
          <AlertCircle className="size-4" />
          <AlertTitle>Trading disabled</AlertTitle>
          <AlertDescription>
            {payload.message ?? 'This account cannot trade. Contact an administrator.'}
          </AlertDescription>
        </Alert>
      ) : payload.paused ? (
        <>
          <Alert>
            <AlertCircle className="size-4" />
            <AlertTitle>Paused</AlertTitle>
            <AlertDescription>
              {payload.resume_hint ??
                'Switch bot mode away from paused to resume market scanning.'}
            </AlertDescription>
          </Alert>
          <p className="text-sm text-muted-foreground">No live opportunities while paused.</p>
        </>
      ) : (
        <div className="rounded-md border border-border overflow-x-auto">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[120px]">Status</TableHead>
                <TableHead>Strategy</TableHead>
                <TableHead>Legs</TableHead>
                <TableHead>Max P/L</TableHead>
                <TableHead>Rec.</TableHead>
                <TableHead>Rationale</TableHead>
                <TableHead className="w-[140px] text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {payload.rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={7} className="text-muted-foreground">
                    No rows yet — runner may be idle or policy unset (
                    <code className="text-xs">OPTITRADE_POLICY_PATH</code>).
                  </TableCell>
                </TableRow>
              ) : (
                payload.rows.map((r) => {
                  const a = rowActions(r)
                  const legStr = (r.legs ?? []).map((l) => l.instrument).join(' / ')
                  return (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-xs">{statusLabel(r.status)}</TableCell>
                      <TableCell className="text-sm">{r.strategy_name}</TableCell>
                      <TableCell className="max-w-[240px] truncate text-xs" title={legStr}>
                        {legStr || '—'}
                      </TableCell>
                      <TableCell className="text-xs whitespace-nowrap">
                        {r.max_profit} / {r.max_loss}
                      </TableCell>
                      <TableCell className="text-xs">{r.recommendation}</TableCell>
                      <TableCell className="text-xs text-muted-foreground max-w-[280px]">
                        {r.rationale}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex flex-wrap justify-end gap-1">
                          {a.open ? (
                            <Button
                              type="button"
                              size="sm"
                              variant="secondary"
                              disabled={busyId === r.id}
                              onClick={() => void runAction(r.id, postOpportunityOpen)}
                            >
                              Open
                            </Button>
                          ) : null}
                          {a.cancel ? (
                            <Button
                              type="button"
                              size="sm"
                              variant="outline"
                              disabled={busyId === r.id}
                              onClick={() => void runAction(r.id, postOpportunityCancel)}
                            >
                              Cancel
                            </Button>
                          ) : null}
                          {a.close ? (
                            <Button
                              type="button"
                              size="sm"
                              variant="outline"
                              disabled={busyId === r.id}
                              onClick={() => void runAction(r.id, postOpportunityClose)}
                            >
                              Close
                            </Button>
                          ) : null}
                        </div>
                      </TableCell>
                    </TableRow>
                  )
                })
              )}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  )
}
