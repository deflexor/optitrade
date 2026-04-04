import { useCallback, useEffect, useState } from 'react'
import { Button } from '@/components/ui/button'
import { getTradingStatus, putTradingMode, type TradingStatus } from '@/api/trading'

const modes = ['manual', 'auto', 'paused'] as const

function modeLabel(m: (typeof modes)[number]): string {
  switch (m) {
    case 'manual':
      return 'Manual'
    case 'auto':
      return 'Auto'
    default:
      return 'Paused'
  }
}

export default function BotModeBar() {
  const [status, setStatus] = useState<TradingStatus | null>(null)
  const [pending, setPending] = useState(false)

  const refresh = useCallback(() => {
    void getTradingStatus()
      .then(setStatus)
      .catch(() => setStatus(null))
  }, [])

  useEffect(() => {
    refresh()
    const t = window.setInterval(refresh, 30_000)
    return () => window.clearInterval(t)
  }, [refresh])

  const setMode = async (m: (typeof modes)[number]) => {
    setPending(true)
    try {
      await putTradingMode(m)
      await refresh()
    } finally {
      setPending(false)
    }
  }

  if (!status) {
    return null
  }

  return (
    <div className="flex flex-wrap items-center gap-2 border border-border rounded-md px-2 py-1.5 bg-muted/30">
      <span className="text-xs font-medium text-muted-foreground uppercase tracking-wide">
        Bot
      </span>
      <div className="flex gap-1" role="group" aria-label="Trading bot mode">
        {modes.map((m) => (
          <Button
            key={m}
            type="button"
            size="sm"
            variant={status.bot_mode === m ? 'default' : 'outline'}
            className="h-8 px-2.5 text-xs"
            disabled={pending}
            onClick={() => void setMode(m)}
          >
            {modeLabel(m)}
          </Button>
        ))}
      </div>
      {!status.runner_running ? (
        <span className="text-xs text-muted-foreground" title="No runner loop for this account">
          Runner idle
        </span>
      ) : (
        <span className="text-xs text-muted-foreground">Runner on</span>
      )}
    </div>
  )
}
