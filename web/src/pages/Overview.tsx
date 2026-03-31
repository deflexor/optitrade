import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import HealthPanel from '../components/HealthPanel'
import RebalanceModal from '../components/RebalanceModal'

type Overview = {
  account: {
    currency: string
    equity: string
    balance: string
    exchange_degraded?: boolean
  }
  pnl_series: {
    points: { t: string; pnl_quote: string }[]
    window: { from: string; to: string }
  }
  market_mood: {
    label: string
    available: boolean
    explanation?: string
  }
  strategy: {
    available: boolean
    message?: string
  }
}

function PnlSparkline({ points }: { points: { t: string; pnl_quote: string }[] }) {
  if (points.length < 2) {
    return (
      <p className="text-sm text-slate-500">Not enough P/L points for this window.</p>
    )
  }
  const vals = points.map((p) => Number(p.pnl_quote))
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const span = max - min || 1
  const w = 400
  const h = 120
  const path = vals
    .map((v, i) => {
      const x = (i / (vals.length - 1)) * w
      const y = h - ((v - min) / span) * (h - 8) - 4
      return `${i === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`
    })
    .join(' ')
  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      className="w-full max-w-xl text-emerald-400/90"
      preserveAspectRatio="none"
      aria-label="P/L series"
    >
      <path d={path} fill="none" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  )
}

export default function Overview() {
  const [data, setData] = useState<Overview | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [rebalance, setRebalance] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const { data: d } = await api.get<Overview>('/overview')
        if (!cancelled) {
          setData(d)
          setErr(null)
        }
      } catch {
        if (!cancelled) {
          setErr('Could not load overview (exchange or session issue).')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <HealthPanel />
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="text-lg font-semibold text-slate-100">Overview</h2>
        <button
          type="button"
          className="rounded-md border border-slate-600 px-3 py-1.5 text-sm text-slate-200 hover:bg-slate-800"
          onClick={() => setRebalance(true)}
        >
          Rebalance…
        </button>
        <Link
          to="/positions"
          className="text-sm text-emerald-400 hover:text-emerald-300"
        >
          Open positions →
        </Link>
      </div>
      {err ? (
        <p className="text-amber-200/90">{err}</p>
      ) : !data ? (
        <p className="text-slate-500">Loading…</p>
      ) : (
        <>
          <section className="rounded-lg border border-slate-800 bg-slate-900/40 p-4">
            <h3 className="text-sm font-medium text-slate-400">Balance</h3>
            <p className="mt-2 font-mono text-2xl text-slate-100">
              {data.account.balance} {data.account.currency}
            </p>
            <p className="text-xs text-slate-500">Equity {data.account.equity}</p>
            {data.account.exchange_degraded ? (
              <p className="mt-2 text-xs text-amber-300/90">
                Exchange data partially degraded — verify fills in your venue UI.
              </p>
            ) : null}
          </section>
          <section className="rounded-lg border border-slate-800 bg-slate-900/40 p-4">
            <h3 className="text-sm font-medium text-slate-400">
              P/L ({data.pnl_series.window.from} → {data.pnl_series.window.to})
            </h3>
            <div className="mt-4">
              <PnlSparkline points={data.pnl_series.points} />
            </div>
          </section>
          <section className="grid gap-4 md:grid-cols-2">
            <div className="rounded-lg border border-slate-800 bg-slate-900/40 p-4">
              <h3 className="text-sm font-medium text-slate-400">Market mood</h3>
              {data.market_mood.available ? (
                <p className="mt-2 text-slate-200">{data.market_mood.label}</p>
              ) : (
                <p className="mt-2 text-sm text-amber-200/80">
                  {data.market_mood.explanation ?? 'Unavailable'}
                </p>
              )}
            </div>
            <div className="rounded-lg border border-slate-800 bg-slate-900/40 p-4">
              <h3 className="text-sm font-medium text-slate-400">Strategy</h3>
              {data.strategy.available ? (
                <p className="mt-2 text-slate-200">Loaded</p>
              ) : (
                <p className="mt-2 text-sm text-amber-200/80">
                  {data.strategy.message ?? 'Unavailable'}
                </p>
              )}
            </div>
          </section>
        </>
      )}
      {rebalance ? (
        <RebalanceModal onClose={() => setRebalance(false)} />
      ) : null}
    </div>
  )
}
