import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
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

export default function PositionsPage() {
  const [open, setOpen] = useState<OpenResp | null>(null)
  const [closed, setClosed] = useState<ClosedResp | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [cursor, setCursor] = useState('')
  const [closeId, setCloseId] = useState<string | null>(null)
  const [closeLabel, setCloseLabel] = useState('')

  async function loadPage(c: string) {
    setErr(null)
    try {
      const q = c ? `?cursor=${encodeURIComponent(c)}` : ''
      const [{ data: o }, { data: cl }] = await Promise.all([
        api.get<OpenResp>(`/positions/open${q}`),
        api.get<ClosedResp>('/positions/closed'),
      ])
      setOpen(o)
      setClosed(cl)
    } catch {
      setErr('Could not load positions.')
    }
  }

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const [{ data: o }, { data: cl }] = await Promise.all([
          api.get<OpenResp>('/positions/open'),
          api.get<ClosedResp>('/positions/closed'),
        ])
        if (!cancelled) {
          setOpen(o)
          setClosed(cl)
          setErr(null)
        }
      } catch {
        if (!cancelled) setErr('Could not load positions.')
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-slate-100">Positions</h2>
        <Link to="/" className="text-sm text-emerald-400 hover:text-emerald-300">
          ← Overview
        </Link>
      </div>
      {err ? <p className="text-amber-200/90">{err}</p> : null}

      <section>
        <h3 className="text-sm font-medium text-slate-400">Open</h3>
        {!open ? (
          <p className="text-slate-500">Loading…</p>
        ) : open.items.length === 0 ? (
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
                {open.items.map((r) => (
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
        {!closed ? (
          <p className="text-slate-500">Loading…</p>
        ) : closed.items.length === 0 ? (
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
                {closed.items.map((r) => (
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
