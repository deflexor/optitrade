import { Outlet, Route, Routes } from 'react-router-dom'

function Shell() {
  return (
    <div className="min-h-screen bg-slate-950 text-slate-100">
      <header className="border-b border-slate-800 bg-slate-900/80 px-6 py-4 backdrop-blur">
        <h1 className="text-lg font-semibold tracking-tight">
          Optitrade Dashboard
        </h1>
      </header>
      <main className="mx-auto max-w-5xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  )
}

function Home() {
  return (
    <p className="text-slate-400">
      Operator UI scaffold. API requests use{' '}
      <code className="rounded bg-slate-800 px-1.5 py-0.5 text-sm text-slate-200">
        /api/v1
      </code>{' '}
      via the Vite dev proxy.
    </p>
  )
}

export default function App() {
  return (
    <Routes>
      <Route element={<Shell />}>
        <Route path="/" element={<Home />} />
      </Route>
    </Routes>
  )
}
