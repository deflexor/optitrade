import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
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
      <Link
        to="/positions"
        className="text-sm text-primary underline-offset-4 hover:underline"
      >
        ← Positions
      </Link>
      {err ? (
        <p className="text-amber-200/90">{err}</p>
      ) : !data ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <>
          <h2 className="text-lg font-semibold text-foreground">{data.instrument_summary}</h2>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">Legs</h3>
            </CardHeader>
            <CardContent>
              <ul className="space-y-1 text-sm text-card-foreground">
                {data.legs.map((l) => (
                  <li key={l.instrument_name} className="font-mono text-xs">
                    {l.instrument_name} · {l.direction ?? '—'} · {l.size ?? '—'}
                  </li>
                ))}
              </ul>
            </CardContent>
          </Card>
          <Separator />
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">Metrics</h3>
            </CardHeader>
            <CardContent>
              <dl className="grid grid-cols-2 gap-2 text-xs text-card-foreground md:grid-cols-3">
                {Object.entries(data.metrics).map(([k, v]) => (
                  <div key={k}>
                    <dt className="text-muted-foreground">{k}</dt>
                    <dd className="font-mono">{v}</dd>
                  </div>
                ))}
              </dl>
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">Greeks</h3>
            </CardHeader>
            <CardContent>
              {data.greeks.available === false ? (
                <p className="text-sm text-amber-200/80">{data.greeks.note ?? 'N/A'}</p>
              ) : (
                <dl className="grid grid-cols-2 gap-2 text-xs text-card-foreground">
                  {data.greeks.delta ? (
                    <div>
                      <dt className="text-muted-foreground">delta</dt>
                      <dd className="font-mono">{data.greeks.delta}</dd>
                    </div>
                  ) : null}
                  {data.greeks.gamma ? (
                    <div>
                      <dt className="text-muted-foreground">gamma</dt>
                      <dd className="font-mono">{data.greeks.gamma}</dd>
                    </div>
                  ) : null}
                </dl>
              )}
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
