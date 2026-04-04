import { useEffect, type ReactNode } from 'react'
import { Link, Outlet, Route, Routes } from 'react-router-dom'
import { LayoutDashboard } from 'lucide-react'
import { Button } from '@/components/ui/button'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import Overview from './pages/Overview'
import PositionDetail from './pages/PositionDetail'
import PositionsPage from './pages/PositionsPage'
import SettingsPage from './pages/SettingsPage'
import { useAuthStore } from './stores/authStore'

function Shell() {
  const { username, logout } = useAuthStore()
  return (
    <div className="min-h-screen bg-background text-foreground">
      <header className="border-b border-border bg-card/80 px-6 py-4 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-4">
          <h1 className="text-lg font-semibold tracking-tight">
            <Link
              to="/"
              className="inline-flex items-center gap-2 text-foreground hover:text-primary"
            >
              <LayoutDashboard className="size-5 text-primary" aria-hidden />
              Optitrade Dashboard
            </Link>
          </h1>
          {username ? (
            <div className="flex items-center gap-3 text-sm text-muted-foreground">
              <Button variant="outline" size="sm" asChild>
                <Link to="/settings">Settings</Link>
              </Button>
              <span className="hidden sm:inline">{username}</span>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => void logout()}
              >
                Sign out
              </Button>
            </div>
          ) : null}
        </div>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}

function Bootstrap({ children }: { children: ReactNode }) {
  const bootstrap = useAuthStore((s) => s.bootstrap)
  useEffect(() => {
    void bootstrap()
  }, [bootstrap])
  return children
}

export default function App() {
  return (
    <Bootstrap>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<Shell />}>
          <Route element={<ProtectedRoute />}>
            <Route path="/" element={<Overview />} />
            <Route path="/positions" element={<PositionsPage />} />
            <Route path="/positions/:id" element={<PositionDetail />} />
            <Route path="/settings" element={<SettingsPage />} />
          </Route>
        </Route>
      </Routes>
    </Bootstrap>
  )
}
