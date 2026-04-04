import { useState } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useAuthStore } from '../stores/authStore'

export default function Login() {
  const { username, login, loginError, clearError } = useAuthStore()
  const loc = useLocation()
  const [u, setU] = useState('')
  const [p, setP] = useState('')
  const [busy, setBusy] = useState(false)

  if (username) {
    const to = (loc.state as { from?: string })?.from ?? '/'
    return <Navigate to={to} replace />
  }

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault()
    clearError()
    setBusy(true)
    try {
      await login(u.trim(), p)
    } catch {
      /* error surfaced via store */
    } finally {
      setBusy(false)
      setP('')
    }
  }

  return (
    <div className="min-h-screen bg-background px-4 py-16 text-foreground">
      <div className="mx-auto max-w-sm">
        <Card className="border-border bg-card">
          <CardHeader>
            <h2 className="text-xl font-semibold tracking-tight">Operator sign-in</h2>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="user">Username</Label>
                <Input
                  id="user"
                  autoComplete="username"
                  value={u}
                  onChange={(e) => setU(e.target.value)}
                  required
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="pass">Password</Label>
                <Input
                  id="pass"
                  type="password"
                  autoComplete="current-password"
                  value={p}
                  onChange={(e) => setP(e.target.value)}
                  required
                />
              </div>
              {loginError ? (
                <Alert variant="destructive">
                  <AlertDescription>{loginError}</AlertDescription>
                </Alert>
              ) : null}
              <Button type="submit" className="w-full" disabled={busy}>
                {busy ? 'Signing in…' : 'Sign in'}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
