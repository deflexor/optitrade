import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api/client'

type Detail = {
  id: string
  instrument_summary: string
  legs: { instrument_name: string; size?: number; direction?: string }[]
  metrics: Record<string, string>
  greeks: { available?: boolean; note?: string; delta?: string; gamma?: string }
}

export default function PositionDetail() {
  const { id = '' } = useParams()
  const [data, setData] = useState<Detail | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const { data: d } = await api.get<Detail>(
          `/positions/${encodeURIComponent(id)}`,
        )
        if (!cancelled) {
          setData(d)
          setErr(null)
        }
      } catch {
        if (!cancelled) setErr('Position not found or feed degraded.')
      }
    })()
    return () => {
      cancelled = true
    }
  }, [id])

  return (
    <div className="space-y-6">
      <Link to="/positions" className="text-sm text-emerald-400 hover:text-emerald-300">
        ← Positions
      </Link>
      {err ? (
        <p className="text-amber-200/90">{err}</p>
      ) : !data ? (
        <p className="text-slate-500">Loading…</p>
      ) : (
        <>
          <h2 className="text-lg font-semibold text-slate-100">
            {data.instrument_summary}
          </h2>
          <section>
            <h3 className="text-sm text-slate-400">Legs</h3>
            <ul className="mt-2 space-y-1 text-sm text-slate-200">
              {data.legs.map((l) => (
                <li key={l.instrument_name} className="font-mono text-xs">
                  {l.instrument_name} · {l.direction ?? '—'} · {l.size ?? '—'}
                </li>
              ))}
            </ul>
          </section>
          <section>
            <h3 className="text-sm text-slate-400">Metrics</h3>
            <dl className="mt-2 grid grid-cols-2 gap-2 text-xs text-slate-300 md:grid-cols-3">
              {Object.entries(data.metrics).map(([k, v]) => (
                <div key={k}>
                  <dt className="text-slate-500">{k}</dt>
                  <dd className="font-mono">{v}</dd>
                </div>
              ))}
            </dl>
          </section>
          <section>
            <h3 className="text-sm text-slate-400">Greeks</h3>
            {data.greeks.available === false ? (
              <p className="mt-2 text-sm text-amber-200/80">
                {data.greeks.note ?? 'N/A'}
              </p>
            ) : (
              <dl className="mt-2 grid grid-cols-2 gap-2 text-xs text-slate-300">
                {data.greeks.delta ? (
                  <div>
                    <dt className="text-slate-500">delta</dt>
                    <dd className="font-mono">{data.greeks.delta}</dd>
                  </div>
                ) : null}
                {data.greeks.gamma ? (
                  <div>
                    <dt className="text-slate-500">gamma</dt>
                    <dd className="font-mono">{data.greeks.gamma}</dd>
                  </div>
                ) : null}
              </dl>
            )}
          </section>
        </>
      )}
    </div>
  )
}
