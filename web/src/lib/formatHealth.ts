/** Formats process uptime for operators (compact day/hour/minute/second segments). */
export function formatUptimeSeconds(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds < 0) {
    return '—'
  }
  let s = Math.floor(totalSeconds)
  if (s < 60) {
    return `${s}s`
  }
  const parts: string[] = []
  if (s >= 86400) {
    parts.push(`${Math.floor(s / 86400)}d`)
    s %= 86400
  }
  if (s >= 3600) {
    parts.push(`${Math.floor(s / 3600)}h`)
    s %= 3600
  }
  if (s >= 60) {
    parts.push(`${Math.floor(s / 60)}m`)
    s %= 60
  }
  if (s > 0 || parts.length === 0) {
    parts.push(`${s}s`)
  }
  return parts.join(' ')
}

/** IEC binary units (KiB, MiB, GiB) for heap display. */
export function formatHeapBytes(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) {
    return '—'
  }
  if (bytes < 1024) {
    return `${Math.round(bytes)} B`
  }
  const units = ['KiB', 'MiB', 'GiB', 'TiB'] as const
  let n = bytes / 1024
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  const rounded = n >= 10 || i === 0 ? Math.round(n) : Math.round(n * 10) / 10
  return `${rounded} ${units[i]}`
}
