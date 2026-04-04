import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
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
      <p className="text-sm text-muted-foreground">
        Not enough P/L points for this window.
      </p>
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
      className="w-full max-w-xl text-primary"
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
        <h2 className="text-lg font-semibold text-foreground">Overview</h2>
        <Button type="button" variant="outline" size="sm" onClick={() => setRebalance(true)}>
          Rebalance…
        </Button>
        <Link
          to="/opportunities"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          Opportunities →
        </Link>
      </div>
      {err ? (
        <p className="text-amber-200/90">{err}</p>
      ) : !data ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">Balance</h3>
            </CardHeader>
            <CardContent>
              <p className="font-mono text-2xl text-card-foreground">
                {data.account.balance} {data.account.currency}
              </p>
              <p className="text-xs text-muted-foreground">Equity {data.account.equity}</p>
              {data.account.exchange_degraded ? (
                <p className="mt-2 text-xs text-amber-300/90">
                  Exchange data partially degraded — verify fills in your venue UI.
                </p>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">
                P/L ({data.pnl_series.window.from} → {data.pnl_series.window.to})
              </h3>
            </CardHeader>
            <CardContent>
              <PnlSparkline points={data.pnl_series.points} />
            </CardContent>
          </Card>
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <h3 className="text-sm font-medium text-muted-foreground">Market mood</h3>
              </CardHeader>
              <CardContent>
                {data.market_mood.available ? (
                  <p className="text-card-foreground">{data.market_mood.label}</p>
                ) : (
                  <p className="text-sm text-amber-200/80">
                    {data.market_mood.explanation ?? 'Unavailable'}
                  </p>
                )}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <h3 className="text-sm font-medium text-muted-foreground">Strategy</h3>
              </CardHeader>
              <CardContent>
                {data.strategy.available ? (
                  <p className="text-card-foreground">Loaded</p>
                ) : (
                  <p className="text-sm text-amber-200/80">
                    {data.strategy.message ?? 'Unavailable'}
                  </p>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}
      {rebalance ? <RebalanceModal onClose={() => setRebalance(false)} /> : null}
    </div>
  )
}
