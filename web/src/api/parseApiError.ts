import axios from 'axios'

export type ParsedApiError = { error: string; message: string }

/** Parses BFF `{ error, message }` JSON from an Axios error response. */
export function parseApiError(err: unknown): ParsedApiError | null {
  if (!axios.isAxiosError(err)) {
    return null
  }
  const data = err.response?.data
  if (!data || typeof data !== 'object') {
    return null
  }
  const rec = data as Record<string, unknown>
  const code = rec.error
  const message = rec.message
  if (typeof code === 'string' && typeof message === 'string') {
    return { error: code, message }
  }
  return null
}
