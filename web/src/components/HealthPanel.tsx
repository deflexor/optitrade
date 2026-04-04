import { useEffect, useState } from 'react'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { api } from '../api/client'
import { formatHeapBytes, formatUptimeSeconds } from '../lib/formatHealth'

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
      <Card className="border-destructive/50 bg-destructive/10">
        <CardContent className="pt-6 text-sm text-destructive">{err}</CardContent>
      </Card>
    )
  }
  if (!data) {
    return (
      <Card>
        <CardContent className="pt-6 text-sm text-muted-foreground">
          Loading health…
        </CardContent>
      </Card>
    )
  }

  const live = data.trading.mode === 'live'
  const modeCardClass = live
    ? 'border-destructive/50 bg-destructive/15 text-destructive-foreground'
    : 'border-primary/50 bg-primary/10 text-primary'

  return (
    <div className="grid gap-3 md:grid-cols-2">
      <Card className={modeCardClass} aria-label="Trading mode">
        <CardHeader className="pb-2">
          <p className="text-xs font-medium uppercase tracking-wide opacity-80">Mode</p>
        </CardHeader>
        <CardContent>
          <p className="text-lg font-semibold capitalize">{data.trading.mode}</p>
          <p className="mt-1 text-xs opacity-90">
            Exchange:{' '}
            {data.trading.exchange_reachable ? 'reachable' : 'unreachable'}
            {data.trading.detail ? ` · ${data.trading.detail}` : ''}
          </p>
        </CardContent>
      </Card>
      <Card>
        <CardHeader className="pb-2">
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Process
          </p>
        </CardHeader>
        <CardContent>
          <p
            className="mt-1 font-mono text-sm text-card-foreground"
            data-testid="health-process-metrics"
          >
            Uptime {formatUptimeSeconds(data.health.uptime_seconds)} · heap{' '}
            {formatHeapBytes(data.health.memory_heap_alloc_bytes)}
          </p>
          <p className="mt-1 text-xs text-muted-foreground">{data.health.collected_at}</p>
        </CardContent>
      </Card>
    </div>
  )
}
