import { useEffect, useState } from 'react'
import { api } from '../api/client'

type HealthResp = {
  health: {
    uptime_seconds: number
    memory_heap_alloc_bytes: number
    collected_at: string
  }
  trading: {
    mode: string
    exchange_reachable: boolean
    detail?: string
  }
}

export default function HealthPanel() {
  const [data, setData] = useState<HealthResp | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const { data: d } = await api.get<HealthResp>('/health')
        if (!cancelled) {
          setData(d)
          setErr(null)
        }
      } catch {
        if (!cancelled) setErr('Health feed unavailable')
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  if (err) {
    return (
      <section className="rounded-lg border border-slate-800 bg-slate-900/50 p-4 text-sm text-amber-200/80">
        {err}
      </section>
    )
  }
  if (!data) {
    return (
      <section className="rounded-lg border border-slate-800 bg-slate-900/50 p-4 text-sm text-slate-500">
        Loading health…
      </section>
    )
  }

  const live = data.trading.mode === 'live'
  const modeStyles = live
    ? 'border-rose-500/50 bg-rose-950/40 text-rose-100'
    : 'border-emerald-500/50 bg-emerald-950/40 text-emerald-100'

  return (
    <section className="grid gap-3 md:grid-cols-2">
      <div
        className={`rounded-lg border px-4 py-3 ${modeStyles}`}
        aria-label="Trading mode"
      >
        <p className="text-xs font-medium uppercase tracking-wide opacity-80">
          Mode
        </p>
        <p className="text-lg font-semibold capitalize">
          {data.trading.mode}
        </p>
        <p className="mt-1 text-xs opacity-90">
          Exchange:{' '}
          {data.trading.exchange_reachable ? 'reachable' : 'unreachable'}
          {data.trading.detail ? ` · ${data.trading.detail}` : ''}
        </p>
      </div>
      <div className="rounded-lg border border-slate-800 bg-slate-900/50 px-4 py-3 text-slate-200">
        <p className="text-xs font-medium uppercase tracking-wide text-slate-500">
          Process
        </p>
        <p className="mt-1 font-mono text-sm">
          uptime {data.health.uptime_seconds}s · heap_alloc{' '}
          {data.health.memory_heap_alloc_bytes} bytes
        </p>
        <p className="mt-1 text-xs text-slate-500">{data.health.collected_at}</p>
      </div>
    </section>
  )
}
