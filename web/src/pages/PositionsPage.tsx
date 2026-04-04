import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { api } from '../api/client'
import { parseApiError } from '../api/parseApiError'
import CloseModal from '../components/CloseModal'

type OpenRow = {
  id: string
  instrument_summary: string
  quote_pnl?: string
  direction?: string
  size?: number
}

type OpenResp = {
  items: OpenRow[]
  next_cursor: string
  total_count: number
}

type ClosedRow = {
  id: string
  instrument_summary: string
  closed_at: string
  realized_pnl_usd: string
  percent_basis_label: string
}

type ClosedResp = {
  items: ClosedRow[]
  truncated: boolean
}

type OpenSlice =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ok'; data: OpenResp }

type ClosedSlice =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ok'; data: ClosedResp }

function positionsErrorMessage(err: unknown): string {
  const p = parseApiError(err)
  if (!p) {
    return 'Could not load positions.'
  }
  if (p.error === 'exchange_unavailable') {
    return 'Trading link is not configured — position data is unavailable. Connect the exchange or try again later.'
  }
  return p.message || 'Could not load positions.'
}

export default function PositionsPage() {
  const [openState, setOpenState] = useState<OpenSlice>({ status: 'loading' })
  const [closedState, setClosedState] = useState<ClosedSlice>({ status: 'loading' })
  const [cursor, setCursor] = useState('')
  const [closeId, setCloseId] = useState<string | null>(null)
  const [closeLabel, setCloseLabel] = useState('')

  async function loadPage(c: string) {
    setOpenState({ status: 'loading' })
    try {
      const q = c ? `?cursor=${encodeURIComponent(c)}` : ''
      const { data: o } = await api.get<OpenResp>(`/positions/open${q}`)
      setOpenState({ status: 'ok', data: o })
    } catch (e) {
      setOpenState({ status: 'error', message: positionsErrorMessage(e) })
    }
  }

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setOpenState({ status: 'loading' })
      setClosedState({ status: 'loading' })
      const openP = api
        .get<OpenResp>('/positions/open')
        .then((r) => ({ ok: true as const, data: r.data }), (e) => ({ ok: false as const, err: e }))
      const closedP = api
        .get<ClosedResp>('/positions/closed')
        .then((r) => ({ ok: true as const, data: r.data }), (e) => ({ ok: false as const, err: e }))
      const [oRes, cRes] = await Promise.all([openP, closedP])
      if (cancelled) {
        return
      }
      if (oRes.ok) {
        setOpenState({ status: 'ok', data: oRes.data })
      } else {
        setOpenState({ status: 'error', message: positionsErrorMessage(oRes.err) })
      }
      if (cRes.ok) {
        setClosedState({ status: 'ok', data: cRes.data })
      } else {
        setClosedState({ status: 'error', message: positionsErrorMessage(cRes.err) })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const open = openState.status === 'ok' ? openState.data : null
  const closed = closedState.status === 'ok' ? closedState.data : null

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-foreground">Positions</h2>
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
      </div>

      <section>
        <Card>
        <CardHeader>
          <h3 className="text-sm font-medium text-muted-foreground">Open</h3>
        </CardHeader>
        <CardContent>
          {openState.status === 'loading' ? (
            <p className="text-muted-foreground">Loading…</p>
          ) : openState.status === 'error' ? (
            <p className="text-amber-200/90">{openState.message}</p>
          ) : openState.data.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No open positions.</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="text-xs uppercase text-muted-foreground">
                      Instrument
                    </TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Dir</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Size</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">U.PnL</TableHead>
                    <TableHead className="text-right text-xs uppercase text-muted-foreground" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {openState.data.items.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-xs">{r.instrument_summary}</TableCell>
                      <TableCell>{r.direction ?? '—'}</TableCell>
                      <TableCell>{r.size ?? '—'}</TableCell>
                      <TableCell className="font-mono">{r.quote_pnl ?? '—'}</TableCell>
                      <TableCell className="space-x-2 text-right">
                        <Button variant="link" className="h-auto p-0 text-primary" asChild>
                          <Link to={`/positions/${encodeURIComponent(r.id)}`}>Detail</Link>
                        </Button>
                        <Button
                          type="button"
                          variant="link"
                          className="h-auto p-0 text-destructive"
                          onClick={() => {
                            setCloseId(r.id)
                            setCloseLabel(r.instrument_summary)
                          }}
                        >
                          Start close
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
          {open && open.next_cursor ? (
            <div className="mt-2 flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  setCursor(open.next_cursor)
                  void loadPage(open.next_cursor)
                }}
              >
                Next page
              </Button>
              {cursor ? (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setCursor('')
                    void loadPage('')
                  }}
                >
                  First page
                </Button>
              ) : null}
            </div>
          ) : null}
          {open ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Showing {open.items.length} of {open.total_count} open (25/page).
            </p>
          ) : null}
        </CardContent>
        </Card>
      </section>

      <section>
        <Card>
        <CardHeader>
          <h3 className="text-sm font-medium text-muted-foreground">
            Closed (30d){' '}
            {closed?.truncated ? (
              <span className="text-amber-200/80">— capped at 200</span>
            ) : null}
          </h3>
        </CardHeader>
        <CardContent>
          {closedState.status === 'loading' ? (
            <p className="text-muted-foreground">Loading…</p>
          ) : closedState.status === 'error' ? (
            <p className="text-amber-200/90">{closedState.message}</p>
          ) : closedState.data.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No closed rows in window.</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="text-xs uppercase text-muted-foreground">
                      Instrument
                    </TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Closed</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Realized</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">% basis</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {closedState.data.items.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-xs">{r.instrument_summary}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{r.closed_at}</TableCell>
                      <TableCell className="font-mono">{r.realized_pnl_usd}</TableCell>
                      <TableCell
                        className="text-xs text-muted-foreground"
                        title={r.percent_basis_label}
                      >
                        {r.percent_basis_label}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
        </Card>
      </section>

      {closeId ? (
        <CloseModal
          positionId={closeId}
          label={closeLabel}
          onClose={() => {
            setCloseId(null)
            void loadPage(cursor)
          }}
        />
      ) : null}
    </div>
  )
}
