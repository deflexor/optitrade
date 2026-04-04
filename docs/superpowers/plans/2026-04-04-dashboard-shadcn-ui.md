# Dashboard shadcn/ui + design tokens — implementation plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add shadcn/ui with a deep-terminal CSS-variable theme and migrate every dashboard screen and modal in `web/` while keeping Playwright e2e behavior stable.

**Architecture:** Vite resolves `@/` to `src/`. Tailwind 3.4 uses `darkMode: 'class'`, semantic colors from `hsl(var(--token))`, and `tailwindcss-animate`. `index.html` sets `<html class="dark">`. Primitives live in `src/components/ui/*` from the shadcn CLI (`new-york` style, CSS variables). Pages compose those primitives; modals use Radix `Dialog` with outside-click dismissed disabled to match today’s UX.

**Tech stack:** React 19, Vite 8, TypeScript 5.9, Tailwind 3.4, shadcn/ui (Radix, CVA, tailwind-merge, clsx), Lucide, Playwright e2e, Go BFF (unchanged).

---

## File map

| Path | Role |
|------|------|
| `web/package.json` / `web/package-lock.json` | New deps from `npm install` + shadcn `add` |
| `web/vite.config.ts` | `resolve.alias['@']` → `src` |
| `web/tsconfig.app.json` | `paths`: `@/*` → `./src/*` |
| `web/components.json` | shadcn project config (created by you or init) |
| `web/tailwind.config.js` | `darkMode`, theme extend, `tailwindcss-animate` plugin |
| `web/src/index.css` | `@tailwind` + `:root` / `.dark` HSL variables + optional `@layer base` |
| `web/index.html` | `<html class="dark" …>` |
| `web/src/lib/utils.ts` | `cn()` helper |
| `web/src/components/ui/*.tsx` | Generated primitives |
| `web/src/App.tsx` | Shell uses `Button`, token utility classes |
| `web/src/main.tsx` | Unchanged imports (or add nothing if theme is html-only) |
| `web/src/pages/*.tsx` | Migrated layouts |
| `web/src/components/HealthPanel.tsx` | Cards, tokens, keep `data-testid` |
| `web/src/components/CloseModal.tsx` | `Dialog`, no backdrop dismiss |
| `web/src/components/RebalanceModal.tsx` | `Dialog`, no backdrop dismiss |

---

### Task 1: Baseline — verify green before UI work

**Files:** None (read-only).

**Test:** Existing Playwright + TypeScript + ESLint gates.

- [ ] **Step 1: Typecheck and lint**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: `npm test` runs `tsc -b` with no errors; ESLint exits 0.

- [ ] **Step 2: End-to-end suite (needs Go toolchain + network for first `go run`)**

Run:

```bash
cd /home/dfr/optitrade/web && npm run test:e2e
```

Expected: all specs pass (dashboard + auth + health + positions + market mood).

- [ ] **Step 3: Commit (only if you created a branch or need a checkpoint)**

If you use a worktree/branch for this feature, an empty commit is optional. Otherwise skip.

```bash
cd /home/dfr/optitrade && git status
```

Expected: clean or only intentional feature-branch changes.

---

### Task 2: Path alias + core npm dependencies

**Files:**

- Modify: `web/vite.config.ts`
- Modify: `web/package.json` / `web/package-lock.json` (via `npm install`)
- Test: `web/` — `npm test`

- [ ] **Step 1: Install runtime styling deps**

Run:

```bash
cd /home/dfr/optitrade/web && npm install class-variance-authority clsx tailwind-merge lucide-react tailwindcss-animate @radix-ui/react-slot
```

Expected: `package.json` lists those packages; `npm install` exits 0.

- [ ] **Step 2: Add Vite alias `@` → `src`**

Replace `web/vite.config.ts` with:

```ts
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

const __dirname = path.dirname(fileURLToPath(import.meta.url))

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, 'src'),
    },
  },
  server: {
    proxy: {
      '/api': {
        target: 'http://127.0.0.1:8080',
        changeOrigin: true,
      },
    },
  },
})
```

- [ ] **Step 3: Add TypeScript paths**

Edit `web/tsconfig.app.json` — inside `compilerOptions`, add:

```json
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
```

Place them after `"jsx": "react-jsx",` (keep trailing commas valid JSON).

- [ ] **Step 4: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test
```

Expected: PASS (no new imports yet; config must stay valid).

- [ ] **Step 5: Commit**

```bash
cd /home/dfr/optitrade && git add web/vite.config.ts web/tsconfig.app.json web/package.json web/package-lock.json && git commit -m "chore(web): add @ alias and shadcn base dependencies"
```

---

### Task 3: `cn` helper + `components.json`

**Files:**

- Create: `web/src/lib/utils.ts`
- Create: `web/components.json`
- Test: `cd web && npm test`

- [ ] **Step 1: Add `cn`**

Create `web/src/lib/utils.ts`:

```ts
import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

- [ ] **Step 2: Add shadcn config**

Create `web/components.json`:

```json
{
  "$schema": "https://ui.shadcn.com/schema.json",
  "style": "new-york",
  "rsc": false,
  "tsx": true,
  "tailwind": {
    "config": "tailwind.config.js",
    "css": "src/index.css",
    "baseColor": "slate",
    "cssVariables": true,
    "prefix": ""
  },
  "aliases": {
    "components": "@/components",
    "utils": "@/lib/utils",
    "ui": "@/components/ui",
    "lib": "@/lib",
    "hooks": "@/hooks"
  },
  "iconLibrary": "lucide"
}
```

- [ ] **Step 3: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/lib/utils.ts web/components.json && git commit -m "chore(web): add cn helper and shadcn components.json"
```

---

### Task 4: Tailwind theme + global CSS tokens + dark root

**Files:**

- Modify: `web/tailwind.config.js`
- Modify: `web/src/index.css`
- Modify: `web/index.html`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Replace Tailwind config**

Replace `web/tailwind.config.js` with:

```js
import tailwindcssAnimate from 'tailwindcss-animate'

/** @type {import('tailwindcss').Config} */
export default {
  darkMode: ['class'],
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        popover: {
          DEFAULT: 'hsl(var(--popover))',
          foreground: 'hsl(var(--popover-foreground))',
        },
        card: {
          DEFAULT: 'hsl(var(--card))',
          foreground: 'hsl(var(--card-foreground))',
        },
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
      keyframes: {
        'accordion-down': {
          from: { height: '0' },
          to: { height: 'var(--radix-accordion-content-height)' },
        },
        'accordion-up': {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: '0' },
        },
      },
      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out',
      },
    },
  },
  plugins: [tailwindcssAnimate],
}
```

- [ ] **Step 2: Replace global CSS**

Replace `web/src/index.css` with:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

@layer base {
  * {
    @apply border-border;
  }
  body {
    @apply bg-background text-foreground antialiased;
    margin: 0;
  }
}

/* Deep terminal — light tokens mirror dark so bare :root is safe; app uses html.dark */
:root {
  --background: 222 47% 5%;
  --foreground: 210 40% 98%;
  --card: 222 47% 7%;
  --card-foreground: 210 40% 98%;
  --popover: 222 47% 7%;
  --popover-foreground: 210 40% 98%;
  --primary: 160 84% 39%;
  --primary-foreground: 144 61% 11%;
  --secondary: 217 19% 16%;
  --secondary-foreground: 210 40% 96%;
  --muted: 217 19% 16%;
  --muted-foreground: 215 16% 57%;
  --accent: 217 19% 18%;
  --accent-foreground: 210 40% 96%;
  --destructive: 0 62% 45%;
  --destructive-foreground: 210 40% 98%;
  --border: 217 19% 18%;
  --input: 217 19% 18%;
  --ring: 160 84% 39%;
  --radius: 0.5rem;
}

.dark {
  --background: 222 47% 5%;
  --foreground: 210 40% 98%;
  --card: 222 47% 7%;
  --card-foreground: 210 40% 98%;
  --popover: 222 47% 7%;
  --popover-foreground: 210 40% 98%;
  --primary: 160 84% 39%;
  --primary-foreground: 144 61% 11%;
  --secondary: 217 19% 16%;
  --secondary-foreground: 210 40% 96%;
  --muted: 217 19% 16%;
  --muted-foreground: 215 16% 57%;
  --accent: 217 19% 18%;
  --accent-foreground: 210 40% 96%;
  --destructive: 0 62% 45%;
  --destructive-foreground: 210 40% 98%;
  --border: 217 19% 18%;
  --input: 217 19% 18%;
  --ring: 160 84% 39%;
  --radius: 0.5rem;
}
```

- [ ] **Step 3: Pin dark mode on `<html>`**

Change the opening tag in `web/index.html` to:

```html
<html lang="en" class="dark">
```

- [ ] **Step 4: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 5: Commit**

```bash
cd /home/dfr/optitrade && git add web/tailwind.config.js web/src/index.css web/index.html && git commit -m "feat(web): tailwind semantic tokens and dark html root"
```

---

### Task 5: Generate shadcn primitives (CLI)

**Files:**

- Create: `web/src/components/ui/*.tsx` (and any Radix deps added to `package.json`)
- Test: `cd web && npm test`

- [ ] **Step 1: Add all required components (non-interactive)**

Run from `web/`:

```bash
cd /home/dfr/optitrade/web && npx shadcn@latest add button card input label table dialog separator badge alert select switch --yes
```

Expected: new files under `web/src/components/ui/` for each component; CLI may install additional `@radix-ui/*` packages. If a component name errors on your CLI version, add it separately with the same `--yes` flag.

- [ ] **Step 2: Typecheck**

Run:

```bash
cd /home/dfr/optitrade/web && npm test
```

Expected: PASS. If the CLI generates code that fails `erasableSyntaxOnly` or `verbatimModuleSyntax`, adjust imports/types in the generated files minimally (prefer fixing one line over rewriting the component).

- [ ] **Step 3: Lint**

Run:

```bash
cd /home/dfr/optitrade/web && npm run lint
```

Expected: exit 0. Fix unused imports in generated files if needed.

- [ ] **Step 4: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/components/ui web/package.json web/package-lock.json && git commit -m "feat(web): add shadcn ui primitives"
```

---

### Task 6: Migrate `App.tsx` shell

**Files:**

- Modify: `web/src/App.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Replace shell with token + shadcn `Button`**

Replace the full contents of `web/src/App.tsx` with:

```tsx
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
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/App.tsx && git commit -m "feat(web): restyle app shell with shadcn Button and tokens"
```

---

### Task 7: Migrate `Login.tsx`

**Files:**

- Modify: `web/src/pages/Login.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Implement login with Card, Label, Input, Button, Alert**

Replace the full contents of `web/src/pages/Login.tsx` with:

```tsx
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
                <Alert variant="destructive" role="alert">
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
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/pages/Login.tsx && git commit -m "feat(web): login page with shadcn form primitives"
```

---

### Task 8: Migrate `HealthPanel.tsx`

**Files:**

- Modify: `web/src/components/HealthPanel.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Use Card + token colors; keep `data-testid`**

Replace the full contents of `web/src/components/HealthPanel.tsx` with:

```tsx
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
        <CardContent className="pt-6 text-sm text-destructive">
          {err}
        </CardContent>
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
          <p className="text-xs font-medium uppercase tracking-wide opacity-80">
            Mode
          </p>
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
          <p className="mt-1 text-xs text-muted-foreground">
            {data.health.collected_at}
          </p>
        </CardContent>
      </Card>
    </div>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/components/HealthPanel.tsx && git commit -m "feat(web): health panel with shadcn Card and tokens"
```

---

### Task 9: Migrate `RebalanceModal.tsx` to `Dialog`

**Files:**

- Modify: `web/src/components/RebalanceModal.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Dialog with outside-click prevented**

Replace the full contents of `web/src/components/RebalanceModal.tsx` with:

```tsx
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '../api/client'

type Preview = {
  preview_token: string
  suggested_adjustments: unknown[]
  disclaimer?: string
}

export default function RebalanceModal({ onClose }: { onClose: () => void }) {
  const [preview, setPreview] = useState<Preview | null>(null)
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)

  async function runPreview() {
    setBusy(true)
    setMsg(null)
    try {
      const { data } = await api.post<Preview>('/rebalance/preview', {})
      setPreview(data)
    } catch {
      setMsg('Preview failed.')
    } finally {
      setBusy(false)
    }
  }

  async function confirm() {
    if (!preview?.preview_token) return
    setBusy(true)
    try {
      await api.post('/rebalance/confirm', {
        preview_token: preview.preview_token,
      })
      setMsg('Acknowledged. No rebalance orders were sent in this build.')
    } catch {
      setMsg('Confirm failed — preview may have expired.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-lg"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
        aria-describedby="rebal-desc"
      >
        <DialogHeader>
          <DialogTitle id="rebal-title">Rebalance</DialogTitle>
          <DialogDescription id="rebal-desc">
            Two-step flow: preview estimates, then confirm. This build may not
            submit exchange orders for rebalance.
          </DialogDescription>
        </DialogHeader>
        {!preview ? (
          <Button type="button" onClick={runPreview} disabled={busy}>
            {busy ? 'Loading…' : 'Run preview'}
          </Button>
        ) : (
          <div className="space-y-3 rounded-md border border-border bg-muted/40 p-3 text-sm">
            <p>
              Suggested legs: {preview.suggested_adjustments.length} (empty in
              v0).
            </p>
            {preview.disclaimer ? (
              <p className="text-xs text-amber-200/90">{preview.disclaimer}</p>
            ) : null}
            <DialogFooter className="gap-2 sm:justify-start">
              <Button
                type="button"
                className="bg-amber-600 text-white hover:bg-amber-500"
                onClick={confirm}
                disabled={busy}
              >
                Confirm
              </Button>
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
            </DialogFooter>
          </div>
        )}
        {msg ? <p className="text-sm text-primary">{msg}</p> : null}
        {!preview ? (
          <Button type="button" variant="ghost" className="px-0" onClick={onClose}>
            Close
          </Button>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/components/RebalanceModal.tsx && git commit -m "feat(web): rebalance flow as shadcn Dialog"
```

---

### Task 10: Migrate `Overview.tsx`

**Files:**

- Modify: `web/src/pages/Overview.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Cards, Button, sparkline uses `text-primary`**

Replace the full contents of `web/src/pages/Overview.tsx` with:

```tsx
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import { api } from '../api/client'
import HealthPanel from '../components/HealthPanel'
import RebalanceModal from '../components/RebalanceModal'

type Overview = {
  account: {
    currency: string
    equity: string
    balance: string
    exchange_degraded?: boolean
  }
  pnl_series: {
    points: { t: string; pnl_quote: string }[]
    window: { from: string; to: string }
  }
  market_mood: {
    label: string
    available: boolean
    explanation?: string
  }
  strategy: {
    available: boolean
    message?: string
  }
}

function PnlSparkline({ points }: { points: { t: string; pnl_quote: string }[] }) {
  if (points.length < 2) {
    return (
      <p className="text-sm text-muted-foreground">
        Not enough P/L points for this window.
      </p>
    )
  }
  const vals = points.map((p) => Number(p.pnl_quote))
  const min = Math.min(...vals)
  const max = Math.max(...vals)
  const span = max - min || 1
  const w = 400
  const h = 120
  const path = vals
    .map((v, i) => {
      const x = (i / (vals.length - 1)) * w
      const y = h - ((v - min) / span) * (h - 8) - 4
      return `${i === 0 ? 'M' : 'L'} ${x.toFixed(1)} ${y.toFixed(1)}`
    })
    .join(' ')
  return (
    <svg
      viewBox={`0 0 ${w} ${h}`}
      className="w-full max-w-xl text-primary"
      preserveAspectRatio="none"
      aria-label="P/L series"
    >
      <path d={path} fill="none" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  )
}

export default function Overview() {
  const [data, setData] = useState<Overview | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [rebalance, setRebalance] = useState(false)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        const { data: d } = await api.get<Overview>('/overview')
        if (!cancelled) {
          setData(d)
          setErr(null)
        }
      } catch {
        if (!cancelled) {
          setErr('Could not load overview (exchange or session issue).')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  return (
    <div className="space-y-8">
      <HealthPanel />
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="text-lg font-semibold text-foreground">Overview</h2>
        <Button type="button" variant="outline" size="sm" onClick={() => setRebalance(true)}>
          Rebalance…
        </Button>
        <Link
          to="/positions"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          Open positions →
        </Link>
      </div>
      {err ? (
        <p className="text-amber-200/90">{err}</p>
      ) : !data ? (
        <p className="text-muted-foreground">Loading…</p>
      ) : (
        <>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">Balance</h3>
            </CardHeader>
            <CardContent>
              <p className="font-mono text-2xl text-card-foreground">
                {data.account.balance} {data.account.currency}
              </p>
              <p className="text-xs text-muted-foreground">Equity {data.account.equity}</p>
              {data.account.exchange_degraded ? (
                <p className="mt-2 text-xs text-amber-300/90">
                  Exchange data partially degraded — verify fills in your venue UI.
                </p>
              ) : null}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <h3 className="text-sm font-medium text-muted-foreground">
                P/L ({data.pnl_series.window.from} → {data.pnl_series.window.to})
              </h3>
            </CardHeader>
            <CardContent>
              <PnlSparkline points={data.pnl_series.points} />
            </CardContent>
          </Card>
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <h3 className="text-sm font-medium text-muted-foreground">Market mood</h3>
              </CardHeader>
              <CardContent>
                {data.market_mood.available ? (
                  <p className="text-card-foreground">{data.market_mood.label}</p>
                ) : (
                  <p className="text-sm text-amber-200/80">
                    {data.market_mood.explanation ?? 'Unavailable'}
                  </p>
                )}
              </CardContent>
            </Card>
            <Card>
              <CardHeader>
                <h3 className="text-sm font-medium text-muted-foreground">Strategy</h3>
              </CardHeader>
              <CardContent>
                {data.strategy.available ? (
                  <p className="text-card-foreground">Loaded</p>
                ) : (
                  <p className="text-sm text-amber-200/80">
                    {data.strategy.message ?? 'Unavailable'}
                  </p>
                )}
              </CardContent>
            </Card>
          </div>
        </>
      )}
      {rebalance ? (
        <RebalanceModal onClose={() => setRebalance(false)} />
      ) : null}
    </div>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/pages/Overview.tsx && git commit -m "feat(web): overview cards and primary sparkline"
```

---

### Task 11: Migrate `CloseModal.tsx` to `Dialog`

**Files:**

- Modify: `web/src/components/CloseModal.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Full Dialog migration with outside-click prevented**

Replace the full contents of `web/src/components/CloseModal.tsx` with:

```tsx
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { api } from '../api/client'

type Preview = {
  preview_token: string
  estimated_exit_pnl: string
  wait_vs_close_guidance: string
  assumptions?: string[]
  labels?: string[]
}

export default function CloseModal({
  positionId,
  label,
  onClose,
}: {
  positionId: string
  label: string
  onClose: () => void
}) {
  const [preview, setPreview] = useState<Preview | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  async function runPreview() {
    setBusy(true)
    setErr(null)
    try {
      const { data } = await api.post<Preview>(
        `/positions/${encodeURIComponent(positionId)}/close/preview`,
        {},
      )
      setPreview(data)
    } catch {
      setErr('Preview failed — position may be stale or exchange unavailable.')
    } finally {
      setBusy(false)
    }
  }

  async function confirm() {
    if (!preview?.preview_token) return
    setBusy(true)
    setErr(null)
    try {
      await api.post(
        `/positions/${encodeURIComponent(positionId)}/close/confirm`,
        { preview_token: preview.preview_token },
      )
      onClose()
    } catch {
      setErr('Confirm rejected — re-preview before trying again.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) onClose()
      }}
    >
      <DialogContent
        className="max-h-[90vh] overflow-y-auto sm:max-w-lg"
        onPointerDownOutside={(e) => e.preventDefault()}
        onInteractOutside={(e) => e.preventDefault()}
        aria-describedby="close-desc"
      >
        <DialogHeader>
          <DialogTitle id="close-title">Close position</DialogTitle>
          <p id="close-desc" className="text-sm text-muted-foreground">
            {label}
          </p>
        </DialogHeader>
        <p className="text-xs text-amber-200/90">
          Estimates only. Live reduce-only market orders can slip; verify size
          and venue state before confirming.
        </p>
        {!preview ? (
          <Button type="button" onClick={runPreview} disabled={busy}>
            {busy ? 'Loading…' : 'Preview close'}
          </Button>
        ) : (
          <div className="space-y-2 rounded-md border border-border bg-muted/40 p-3 text-sm">
            <p>
              Est. exit P/L (floating){' '}
              <span className="font-mono">{preview.estimated_exit_pnl}</span>
            </p>
            <p className="text-xs text-muted-foreground">
              {preview.wait_vs_close_guidance}
            </p>
            {preview.labels?.length ? (
              <ul className="list-inside list-disc text-xs text-amber-200/80">
                {preview.labels.map((l) => (
                  <li key={l}>{l}</li>
                ))}
              </ul>
            ) : null}
            <DialogFooter className="gap-2 sm:justify-start">
              <Button type="button" variant="destructive" onClick={confirm} disabled={busy}>
                Confirm close
              </Button>
              <Button type="button" variant="outline" onClick={onClose}>
                Cancel
              </Button>
            </DialogFooter>
          </div>
        )}
        {err ? <p className="text-sm text-destructive">{err}</p> : null}
        {!preview ? (
          <Button type="button" variant="ghost" className="px-0" onClick={onClose}>
            Close
          </Button>
        ) : null}
      </DialogContent>
    </Dialog>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/components/CloseModal.tsx && git commit -m "feat(web): close position modal as shadcn Dialog"
```

---

### Task 12: Migrate `PositionsPage.tsx`

**Files:**

- Modify: `web/src/pages/PositionsPage.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Replace page with Table + Button + Card sections**

Replace the full contents of `web/src/pages/PositionsPage.tsx` with:

```tsx
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { api } from '../api/client'
import { parseApiError } from '../api/parseApiError'
import CloseModal from '../components/CloseModal'

type OpenRow = {
  id: string
  instrument_summary: string
  quote_pnl?: string
  direction?: string
  size?: number
}

type OpenResp = {
  items: OpenRow[]
  next_cursor: string
  total_count: number
}

type ClosedRow = {
  id: string
  instrument_summary: string
  closed_at: string
  realized_pnl_usd: string
  percent_basis_label: string
}

type ClosedResp = {
  items: ClosedRow[]
  truncated: boolean
}

type OpenSlice =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ok'; data: OpenResp }

type ClosedSlice =
  | { status: 'loading' }
  | { status: 'error'; message: string }
  | { status: 'ok'; data: ClosedResp }

function positionsErrorMessage(err: unknown): string {
  const p = parseApiError(err)
  if (!p) {
    return 'Could not load positions.'
  }
  if (p.error === 'exchange_unavailable') {
    return 'Trading link is not configured — position data is unavailable. Connect the exchange or try again later.'
  }
  return p.message || 'Could not load positions.'
}

export default function PositionsPage() {
  const [openState, setOpenState] = useState<OpenSlice>({ status: 'loading' })
  const [closedState, setClosedState] = useState<ClosedSlice>({ status: 'loading' })
  const [cursor, setCursor] = useState('')
  const [closeId, setCloseId] = useState<string | null>(null)
  const [closeLabel, setCloseLabel] = useState('')

  async function loadPage(c: string) {
    setOpenState({ status: 'loading' })
    try {
      const q = c ? `?cursor=${encodeURIComponent(c)}` : ''
      const { data: o } = await api.get<OpenResp>(`/positions/open${q}`)
      setOpenState({ status: 'ok', data: o })
    } catch (e) {
      setOpenState({ status: 'error', message: positionsErrorMessage(e) })
    }
  }

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      setOpenState({ status: 'loading' })
      setClosedState({ status: 'loading' })
      const openP = api
        .get<OpenResp>('/positions/open')
        .then((r) => ({ ok: true as const, data: r.data }), (e) => ({ ok: false as const, err: e }))
      const closedP = api
        .get<ClosedResp>('/positions/closed')
        .then((r) => ({ ok: true as const, data: r.data }), (e) => ({ ok: false as const, err: e }))
      const [oRes, cRes] = await Promise.all([openP, closedP])
      if (cancelled) {
        return
      }
      if (oRes.ok) {
        setOpenState({ status: 'ok', data: oRes.data })
      } else {
        setOpenState({ status: 'error', message: positionsErrorMessage(oRes.err) })
      }
      if (cRes.ok) {
        setClosedState({ status: 'ok', data: cRes.data })
      } else {
        setClosedState({ status: 'error', message: positionsErrorMessage(cRes.err) })
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const open = openState.status === 'ok' ? openState.data : null
  const closed = closedState.status === 'ok' ? closedState.data : null

  return (
    <div className="space-y-8">
      <div className="flex items-center gap-4">
        <h2 className="text-lg font-semibold text-foreground">Positions</h2>
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
      </div>

      <Card>
        <CardHeader>
          <h3 className="text-sm font-medium text-muted-foreground">Open</h3>
        </CardHeader>
        <CardContent>
          {openState.status === 'loading' ? (
            <p className="text-muted-foreground">Loading…</p>
          ) : openState.status === 'error' ? (
            <p className="text-amber-200/90">{openState.message}</p>
          ) : openState.data.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No open positions.</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="text-xs uppercase text-muted-foreground">
                      Instrument
                    </TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Dir</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Size</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">U.PnL</TableHead>
                    <TableHead className="text-right text-xs uppercase text-muted-foreground" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {openState.data.items.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-xs">{r.instrument_summary}</TableCell>
                      <TableCell>{r.direction ?? '—'}</TableCell>
                      <TableCell>{r.size ?? '—'}</TableCell>
                      <TableCell className="font-mono">{r.quote_pnl ?? '—'}</TableCell>
                      <TableCell className="space-x-2 text-right">
                        <Button variant="link" className="h-auto p-0 text-primary" asChild>
                          <Link to={`/positions/${encodeURIComponent(r.id)}`}>Detail</Link>
                        </Button>
                        <Button
                          type="button"
                          variant="link"
                          className="h-auto p-0 text-destructive"
                          onClick={() => {
                            setCloseId(r.id)
                            setCloseLabel(r.instrument_summary)
                          }}
                        >
                          Start close
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
          {open && open.next_cursor ? (
            <div className="mt-2 flex gap-2">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => {
                  setCursor(open.next_cursor)
                  void loadPage(open.next_cursor)
                }}
              >
                Next page
              </Button>
              {cursor ? (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setCursor('')
                    void loadPage('')
                  }}
                >
                  First page
                </Button>
              ) : null}
            </div>
          ) : null}
          {open ? (
            <p className="mt-1 text-xs text-muted-foreground">
              Showing {open.items.length} of {open.total_count} open (25/page).
            </p>
          ) : null}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <h3 className="text-sm font-medium text-muted-foreground">
            Closed (30d){' '}
            {closed?.truncated ? (
              <span className="text-amber-200/80">— capped at 200</span>
            ) : null}
          </h3>
        </CardHeader>
        <CardContent>
          {closedState.status === 'loading' ? (
            <p className="text-muted-foreground">Loading…</p>
          ) : closedState.status === 'error' ? (
            <p className="text-amber-200/90">{closedState.message}</p>
          ) : closedState.data.items.length === 0 ? (
            <p className="text-sm text-muted-foreground">No closed rows in window.</p>
          ) : (
            <div className="overflow-x-auto rounded-md border border-border">
              <Table>
                <TableHeader>
                  <TableRow className="bg-muted/50 hover:bg-muted/50">
                    <TableHead className="text-xs uppercase text-muted-foreground">
                      Instrument
                    </TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Closed</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">Realized</TableHead>
                    <TableHead className="text-xs uppercase text-muted-foreground">% basis</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {closedState.data.items.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell className="font-mono text-xs">{r.instrument_summary}</TableCell>
                      <TableCell className="text-xs text-muted-foreground">{r.closed_at}</TableCell>
                      <TableCell className="font-mono">{r.realized_pnl_usd}</TableCell>
                      <TableCell
                        className="text-xs text-muted-foreground"
                        title={r.percent_basis_label}
                      >
                        {r.percent_basis_label}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          )}
        </CardContent>
      </Card>

      {closeId ? (
        <CloseModal
          positionId={closeId}
          label={closeLabel}
          onClose={() => {
            setCloseId(null)
            void loadPage(cursor)
          }}
        />
      ) : null}
    </div>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/pages/PositionsPage.tsx && git commit -m "feat(web): positions tables with shadcn Table and Card"
```

---

### Task 13: Migrate `PositionDetail.tsx`

**Files:**

- Modify: `web/src/pages/PositionDetail.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Replace with Cards + token link**

Replace the full contents of `web/src/pages/PositionDetail.tsx` with:

```tsx
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
          <h2 className="text-lg font-semibold text-foreground">
            {data.instrument_summary}
          </h2>
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
                <p className="text-sm text-amber-200/80">
                  {data.greeks.note ?? 'N/A'}
                </p>
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
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/pages/PositionDetail.tsx && git commit -m "feat(web): position detail with shadcn Card"
```

---

### Task 14: Migrate `SettingsPage.tsx`

**Files:**

- Modify: `web/src/pages/SettingsPage.tsx`
- Test: `cd web && npm test && npm run lint`

- [ ] **Step 1: Replace with shadcn form primitives**

Replace the full contents of `web/src/pages/SettingsPage.tsx` with:

```tsx
import { useCallback, useEffect, useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Switch } from '@/components/ui/switch'
import { api } from '../api/client'

type FieldDef = {
  id: string
  label: string
  help: string
  kind: string
  required: boolean
  options?: { value: string; label: string }[]
}

type SecretView = { masked: string; configured: boolean }

type SettingsResp = {
  fields: FieldDef[]
  values: Record<string, unknown>
  warnings: string[]
}

export default function SettingsPage() {
  const [payload, setPayload] = useState<SettingsResp | null>(null)
  const [err, setErr] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [form, setForm] = useState<{
    provider: string
    deribit_use_mainnet: boolean
    okx_demo: boolean
    currencies: string
    deribit_client_id: string
    deribit_client_secret: string
    okx_api_key: string
    okx_secret_key: string
    okx_passphrase: string
  } | null>(null)

  const load = useCallback(async () => {
    setErr(null)
    try {
      const { data } = await api.get<SettingsResp>('/settings')
      setPayload(data)
      const v = data.values
      const sec = (k: string) =>
        (v[k] as SecretView | undefined)?.configured
          ? ''
          : ((v[k] as SecretView | undefined)?.masked ?? '')
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        deribit_client_id: sec('deribit_client_id'),
        deribit_client_secret: sec('deribit_client_secret'),
        okx_api_key: sec('okx_api_key'),
        okx_secret_key: sec('okx_secret_key'),
        okx_passphrase: sec('okx_passphrase'),
      })
    } catch {
      setErr('Could not load settings.')
    }
  }, [])

  useEffect(() => {
    void load()
  }, [load])

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (!form) return
    setSaving(true)
    setErr(null)
    try {
      const secrets: Record<string, string> = {}
      if (form.deribit_client_id.trim())
        secrets.deribit_client_id = form.deribit_client_id.trim()
      if (form.deribit_client_secret.trim())
        secrets.deribit_client_secret = form.deribit_client_secret.trim()
      if (form.okx_api_key.trim()) secrets.okx_api_key = form.okx_api_key.trim()
      if (form.okx_secret_key.trim()) secrets.okx_secret_key = form.okx_secret_key.trim()
      if (form.okx_passphrase.trim()) secrets.okx_passphrase = form.okx_passphrase.trim()

      const { data } = await api.put<SettingsResp>('/settings', {
        provider: form.provider,
        deribit_use_mainnet: form.deribit_use_mainnet,
        okx_demo: form.okx_demo,
        currencies: form.currencies.trim(),
        secrets,
      })
      setPayload(data)
      const v = data.values
      const sec = (k: string) =>
        (v[k] as SecretView | undefined)?.configured
          ? ''
          : ((v[k] as SecretView | undefined)?.masked ?? '')
      setForm({
        provider: String(v.provider ?? 'deribit'),
        deribit_use_mainnet: Boolean(v.deribit_use_mainnet),
        okx_demo: Boolean(v.okx_demo),
        currencies: String(v.currencies ?? ''),
        deribit_client_id: sec('deribit_client_id'),
        deribit_client_secret: sec('deribit_client_secret'),
        okx_api_key: sec('okx_api_key'),
        okx_secret_key: sec('okx_secret_key'),
        okx_passphrase: sec('okx_passphrase'),
      })
    } catch (e: unknown) {
      const ax = e as { response?: { data?: { message?: string } } }
      setErr(ax.response?.data?.message ?? 'Save failed.')
    } finally {
      setSaving(false)
    }
  }

  if (!form || !payload) {
    return (
      <div className="space-y-4">
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
        <p className="text-muted-foreground">{err ?? 'Loading…'}</p>
      </div>
    )
  }

  const prov = form.provider

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-foreground">Settings</h2>
        <Link
          to="/"
          className="text-sm text-primary underline-offset-4 hover:underline"
        >
          ← Overview
        </Link>
      </div>

      {payload.warnings.length > 0 ? (
        <Alert
          variant="default"
          className="border-amber-500/40 bg-amber-950/30 text-amber-100"
          role="alert"
        >
          <AlertDescription className="space-y-2 text-amber-100">
            {payload.warnings.map((w) => (
              <p key={w}>{w}</p>
            ))}
          </AlertDescription>
        </Alert>
      ) : null}

      {err ? (
        <p className="text-sm text-destructive" role="alert">
          {err}
        </p>
      ) : null}

      <form onSubmit={(e) => void onSubmit(e)} className="max-w-xl space-y-6">
        <div className="space-y-2">
          <Label>Exchange</Label>
          <Select
            value={prov}
            onValueChange={(value) => setForm({ ...form, provider: value })}
          >
            <SelectTrigger>
              <SelectValue placeholder="Exchange" />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="deribit">Deribit</SelectItem>
              <SelectItem value="okx">OKX</SelectItem>
            </SelectContent>
          </Select>
          <p className="text-xs text-muted-foreground">Venue for this operator account.</p>
        </div>

        <Separator />

        {prov === 'deribit' ? (
          <div className="space-y-4 rounded-lg border border-border bg-card/40 p-4">
            <div className="flex items-center gap-2">
              <Switch
                id="deribit-mainnet"
                checked={form.deribit_use_mainnet}
                onCheckedChange={(checked) =>
                  setForm({ ...form, deribit_use_mainnet: checked })
                }
              />
              <Label htmlFor="deribit-mainnet" className="text-sm font-normal">
                Use Deribit mainnet (off = testnet)
              </Label>
            </div>
            <div className="space-y-2">
              <Label htmlFor="deribit-client-id">Client ID</Label>
              <Input
                id="deribit-client-id"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.deribit_client_id}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, deribit_client_id: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="deribit-client-secret">Client secret</Label>
              <Input
                id="deribit-client-secret"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.deribit_client_secret}
                placeholder="Leave blank to keep saved value"
                onChange={(e) =>
                  setForm({ ...form, deribit_client_secret: e.target.value })
                }
              />
            </div>
          </div>
        ) : (
          <div className="space-y-4 rounded-lg border border-border bg-card/40 p-4">
            <div className="flex items-center gap-2">
              <Switch
                id="okx-demo"
                checked={form.okx_demo}
                onCheckedChange={(checked) => setForm({ ...form, okx_demo: checked })}
              />
              <Label htmlFor="okx-demo" className="text-sm font-normal">
                OKX demo trading (Demo Trading API keys + simulated header)
              </Label>
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-api-key">API key</Label>
              <Input
                id="okx-api-key"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_api_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_api_key: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-secret-key">Secret key</Label>
              <Input
                id="okx-secret-key"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_secret_key}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_secret_key: e.target.value })}
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="okx-passphrase">Passphrase</Label>
              <Input
                id="okx-passphrase"
                type="password"
                autoComplete="off"
                className="font-mono text-sm"
                value={form.okx_passphrase}
                placeholder="Leave blank to keep saved value"
                onChange={(e) => setForm({ ...form, okx_passphrase: e.target.value })}
              />
            </div>
          </div>
        )}

        <Separator />

        <div className="space-y-2">
          <Label htmlFor="currencies">Overview currencies</Label>
          <Input
            id="currencies"
            value={form.currencies}
            placeholder="e.g. BTC, ETH"
            onChange={(e) => setForm({ ...form, currencies: e.target.value })}
          />
          <p className="text-xs text-muted-foreground">
            Comma-separated; blank uses server default (env OPTITRADE_DASHBOARD_CURRENCIES or BTC, ETH).
          </p>
        </div>

        <Button type="submit" disabled={saving}>
          {saving ? 'Saving…' : 'Save settings'}
        </Button>
      </form>
    </div>
  )
}
```

- [ ] **Step 2: Verify**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 3: Commit**

```bash
cd /home/dfr/optitrade && git add web/src/pages/SettingsPage.tsx && git commit -m "feat(web): settings form with shadcn Select and Switch"
```

---

### Task 15: Final verification and regression sweep

**Files:** None (commands only).

- [ ] **Step 1: Typecheck and lint**

Run:

```bash
cd /home/dfr/optitrade/web && npm test && npm run lint
```

Expected: PASS.

- [ ] **Step 2: Playwright**

Run:

```bash
cd /home/dfr/optitrade/web && npm run test:e2e
```

Expected: PASS. If a heading or link assertion fails, adjust the React tree (prefer restoring the exact accessible name) before changing `web/e2e/*.spec.ts`.

- [ ] **Step 3: Final commit (if any fixes)**

Only if Step 2 required code changes:

```bash
cd /home/dfr/optitrade && git add -A && git commit -m "fix(web): align UI with e2e selectors after shadcn migration"
```

- [ ] **Step 4: Push (session policy)**

```bash
cd /home/dfr/optitrade && git pull --rebase && bd dolt push && git push && git status
```

Expected: `git status` shows branch up to date with `origin`.

---

## Plan self-review (spec coverage)

| Spec requirement | Task(s) |
|------------------|---------|
| shadcn + CSS variables + Tailwind semantic mapping | 2–5, 4 |
| `@/` alias | 2 |
| `html` / dark class | 4 |
| Deep terminal palette | 4 (`index.css`) |
| Components: Card, Button, Input, Label, Table, Dialog, Separator, Badge, Alert, Select, Switch | 5 (add); Badge unused in snippets—**add** still OK for spec; optional remove in Task 5 if tree-shaking not a concern |
| Lucide | 6 (`App.tsx` icon) |
| Full page sweep | 6–14 |
| Dialog: no backdrop dismiss | 9, 11 (`onPointerDownOutside` / `onInteractOutside`) |
| Preserve e2e headings, links, alerts, `data-testid` | all migration tasks; HealthPanel Task 8 |
| `npm test`, `npm run lint`, Playwright | 1, 15 |

**Gap closed:** Spec lists **Badge**; plan adds it via CLI (Task 5) but no page uses it yet—acceptable per YAGNI for future chips; spec said “where they clarify” so unused import files are OK.

**Headings:** shadcn `CardTitle` renders as a `div`, which breaks `getByRole('heading', …)` in Playwright. Section titles that e2e asserts on use **`h2` / `h3`** inside `CardHeader` instead of `CardTitle` (Login, Overview, Positions, Position detail).

**Placeholder scan:** No TBD/TODO/similar steps.

**Type consistency:** `SettingsResp` / `form` state matches original Settings page behavior.

---

**Plan complete and saved to `docs/superpowers/plans/2026-04-04-dashboard-shadcn-ui.md`. Two execution options:**

1. **Subagent-driven (recommended)** — Dispatch a fresh subagent per task, review between tasks, fast iteration. **Required sub-skill:** `superpowers:subagent-driven-development`.

2. **Inline execution** — Run tasks in this session using **`superpowers:executing-plans`**, batch execution with checkpoints.

**Which approach do you want?**
