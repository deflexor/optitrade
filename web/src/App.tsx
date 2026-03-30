import { useEffect, type ReactNode } from 'react'
import { Link, Outlet, Route, Routes } from 'react-router-dom'
import ProtectedRoute from './components/ProtectedRoute'
import Login from './pages/Login'
import Overview from './pages/Overview'
import PositionDetail from './pages/PositionDetail'
import PositionsPage from './pages/PositionsPage'
import { useAuthStore } from './stores/authStore'

function Shell() {
  const { username, logout } = useAuthStore()
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900/80 px-6 py-4 backdrop-blur">
        <div className="mx-auto flex max-w-5xl items-center justify-between gap-4">
          <h1 className="text-lg font-semibold tracking-tight">
            <Link to="/" className="hover:text-emerald-300">
              Optitrade Dashboard
            </Link>
          </h1>
          {username ? (
            <div className="flex items-center gap-3 text-sm text-slate-400">
              <span className="hidden sm:inline">{username}</span>
              <button
                type="button"
                className="rounded-md border border-slate-700 px-3 py-1.5 text-slate-200 hover:bg-slate-800"
                onClick={() => void logout()}
              >
                Sign out
              </button>
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
          </Route>
        </Route>
      </Routes>
    </Bootstrap>
  )
}
