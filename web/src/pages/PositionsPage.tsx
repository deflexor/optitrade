import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
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

  const open =
    openState.status === 'ok' ? openState.data : null
  const closed =
    closedState.status === 'ok' ? closedState.data : null

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-slate-100">Positions</h2>
        <Link to="/" className="text-sm text-emerald-400 hover:text-emerald-300">
          ← Overview
        </Link>
      </div>

      <section>
        <h3 className="text-sm font-medium text-slate-400">Open</h3>
        {openState.status === 'loading' ? (
          <p className="text-slate-500">Loading…</p>
        ) : openState.status === 'error' ? (
          <p className="mt-2 text-amber-200/90">{openState.message}</p>
        ) : openState.data.items.length === 0 ? (
          <p className="text-sm text-slate-500">No open positions.</p>
        ) : (
          <div className="mt-2 overflow-x-auto rounded-lg border border-slate-800">
            <table className="min-w-full text-left text-sm text-slate-200">
              <thead className="border-b border-slate-800 bg-slate-900/80 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-3 py-2">Instrument</th>
                  <th className="px-3 py-2">Dir</th>
                  <th className="px-3 py-2">Size</th>
                  <th className="px-3 py-2">U.PnL</th>
                  <th className="px-3 py-2" />
                </tr>
              </thead>
              <tbody>
                {openState.data.items.map((r) => (
                  <tr key={r.id} className="border-b border-slate-800/80">
                    <td className="px-3 py-2 font-mono text-xs">{r.instrument_summary}</td>
                    <td className="px-3 py-2">{r.direction ?? '—'}</td>
                    <td className="px-3 py-2">{r.size ?? '—'}</td>
                    <td className="px-3 py-2 font-mono">{r.quote_pnl ?? '—'}</td>
                    <td className="space-x-2 px-3 py-2 text-right">
                      <Link
                        to={`/positions/${encodeURIComponent(r.id)}`}
                        className="text-emerald-400 hover:text-emerald-300"
                      >
                        Detail
                      </Link>
                      <button
                        type="button"
                        className="text-rose-300 hover:text-rose-200"
                        onClick={() => {
                          setCloseId(r.id)
                          setCloseLabel(r.instrument_summary)
                        }}
                      >
                        Start close
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
        {open && open.next_cursor ? (
          <div className="mt-2 flex gap-2">
            <button
              type="button"
              className="rounded border border-slate-700 px-3 py-1 text-sm text-slate-300 hover:bg-slate-800"
              onClick={() => {
                setCursor(open.next_cursor)
                void loadPage(open.next_cursor)
              }}
            >
              Next page
            </button>
            {cursor ? (
              <button
                type="button"
                className="rounded border border-slate-700 px-3 py-1 text-sm text-slate-300 hover:bg-slate-800"
                onClick={() => {
                  setCursor('')
                  void loadPage('')
                }}
              >
                First page
              </button>
            ) : null}
          </div>
        ) : null}
        {open ? (
          <p className="mt-1 text-xs text-slate-500">
            Showing {open.items.length} of {open.total_count} open (25/page).
          </p>
        ) : null}
      </section>

      <section>
        <h3 className="text-sm font-medium text-slate-400">
          Closed (30d){' '}
          {closed?.truncated ? (
            <span className="text-amber-200/80">— capped at 200</span>
          ) : null}
        </h3>
        {closedState.status === 'loading' ? (
          <p className="text-slate-500">Loading…</p>
        ) : closedState.status === 'error' ? (
          <p className="mt-2 text-amber-200/90">{closedState.message}</p>
        ) : closedState.data.items.length === 0 ? (
          <p className="text-sm text-slate-500">No closed rows in window.</p>
        ) : (
          <div className="mt-2 overflow-x-auto rounded-lg border border-slate-800">
            <table className="min-w-full text-left text-sm text-slate-200">
              <thead className="border-b border-slate-800 bg-slate-900/80 text-xs uppercase text-slate-500">
                <tr>
                  <th className="px-3 py-2">Instrument</th>
                  <th className="px-3 py-2">Closed</th>
                  <th className="px-3 py-2">Realized</th>
                  <th className="px-3 py-2">% basis</th>
                </tr>
              </thead>
              <tbody>
                {closedState.data.items.map((r) => (
                  <tr key={r.id} className="border-b border-slate-800/80">
                    <td className="px-3 py-2 font-mono text-xs">{r.instrument_summary}</td>
                    <td className="px-3 py-2 text-xs text-slate-400">{r.closed_at}</td>
                    <td className="px-3 py-2 font-mono">{r.realized_pnl_usd}</td>
                    <td className="px-3 py-2 text-xs text-slate-400" title={r.percent_basis_label}>
                      {r.percent_basis_label}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
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
