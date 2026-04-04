# Dashboard UI: shadcn/ui, design tokens, full sweep

**Status:** Approved for implementation planning  
**Scope:** `web/` (React 19 + Vite 8 + Tailwind 3.4)  
**Date:** 2026-04-04

## Summary

Upgrade the operator dashboard to a cohesive **deep terminal** visual system: very dark surfaces, crisp borders, **emerald** primary actions and links, **amber** warnings, **rose** for live mode and destructive actions. Introduce **shadcn/ui** primitives (Radix + CVA + tailwind-merge), **global CSS variables** mapped through Tailwind, **Lucide** icons where they aid recognition, and migrate **all** current pages and modals in **one** implementation pass. **No light theme** in this version: the app root stays in dark token mode only.

## Brainstorming decisions

| Topic | Choice |
|--------|--------|
| Overall mood | **A** — Evolve current slate + emerald operator-console aesthetic |
| Rollout | **2** — Foundation (tokens + shadcn) then migrate every screen in one sweep |
| Icons | **1** — Lucide (shadcn default ecosystem) |

## Approach (selected)

Use **shadcn/ui** with the standard Vite + React setup (`components.json`, `src/components/ui/*`), **CSS variables** for theming, and `tailwind.config` extended per shadcn conventions (`hsl(var(--…))`, `tailwindcss-animate`). Alternatives considered: Radix-only (more hand-rolled polish) and tokens-only without shadcn (rejected—does not meet the shadcn requirement).

## Tooling

- **shadcn/ui** init for `web/` with **CSS variables** for colors; `baseColor` aligned with slate.
- **Path alias:** Add `@/` → `src/` in **Vite** (`resolve.alias`) and **TypeScript** (`tsconfig.app.json` `paths`) so imports match shadcn defaults (`@/components/ui/...`).
- **Dependencies** (as required by chosen components): e.g. `tailwindcss-animate`, `class-variance-authority`, `clsx`, `tailwind-merge`, `lucide-react`, and per-component `@radix-ui/*` packages.
- **Tailwind:** Keep **v3.4** unless a required upgrade is unavoidable in the same change; set `darkMode: ['class']`, extend theme for shadcn tokens, add `tailwindcss-animate` to `plugins`.
- **`src/index.css`:** Retain `@tailwind` directives; add shadcn **`:root`** and **`.dark`** variable blocks customized for this palette; optional `@layer base` rules for `body` (background, foreground, font smoothing).

## Global palette (semantic tokens)

Define CSS custom properties consumed by shadcn and Tailwind extensions:

- **Layout:** `--background`, `--foreground`, `--card`, `--card-foreground`, `--popover`, `--popover-foreground`
- **Chrome:** `--border`, `--input`, `--ring` (emerald-tinted focus)
- **Actions:** `--primary`, `--primary-foreground`, `--secondary`, `--secondary-foreground`, `--accent`, `--accent-foreground`
- **Feedback:** `--destructive`, `--destructive-foreground`, `--muted`, `--muted-foreground`
- **Shape:** `--radius` (consistent with current rounded-md / rounded-lg usage)

**Runtime theme:** Set **`class="dark"`** on the document root (e.g. in `main.tsx`) so only the dark variable set applies—no light-mode toggle in v1.

Tune HSL values so the result matches today’s intent: **slate-950/900** surfaces, **emerald** primary, **amber** warnings, **rose** destructive / live emphasis.

## shadcn components (initial set)

Add only primitives needed for the sweep:

| Use | Components |
|-----|------------|
| Layout / sections | `Card` (+ header/title/content as used) |
| Forms | `Button`, `Input`, `Label`, `Select`, `Switch`, `Separator` |
| Data | `Table` |
| Overlays | `Dialog` |
| Feedback | `Badge`, `Alert` (where they clarify errors/warnings) |

## Per-screen mapping

- **`App.tsx`:** Header layout unchanged in structure; **Button** (`outline` / `ghost`) for Settings and Sign out; title link uses token-based colors; **Lucide** optional for nav affordances.
- **`Login.tsx`:** Optional **Card** framing; **Label** + **Input**; primary **Button**; sign-in error keeps **`role="alert"`** for Playwright (via **Alert** or explicit role on the error node).
- **`HealthPanel.tsx`:** **Card** grid; live vs paper styling via semantic tokens (destructive-tinted vs primary-tinted); preserve **`data-testid="health-process-metrics"`** on the process metrics line.
- **`Overview.tsx`:** Metric blocks as **Card**; **Button** for Rebalance; P/L sparkline remains SVG; stroke/color via `text-primary` or CSS variable on container; **RebalanceModal** → **Dialog**.
- **`PositionsPage.tsx`:** **Card** sections; open/closed data via **Table**; pagination **Button** `outline`; row actions as link-styled or ghost **Button**; **CloseModal** → **Dialog**.
- **`PositionDetail.tsx`:** Back link with token colors; sections as **Card** and/or **Separator**; **font-mono** on numeric/instrument values.
- **`SettingsPage.tsx`:** **Label** + **Input** / **Select** / **Switch**; **Separator** between groups; submit **Button**; API warnings can use **Alert** if helpful.

## Modals and interaction parity

Today’s modals **do not** close when clicking the backdrop—only explicit **Cancel** / **Close** actions. When using **Dialog**, prevent default outside-dismiss (e.g. `onInteractOutside` / `onPointerDownOutside` with `preventDefault()`) so behavior matches the current implementation unless product explicitly changes it.

## Error handling and copy

Preserve **wording** that **Playwright** and operators rely on unless a test update is unavoidable; prefer adjusting layout/components while keeping accessible names stable.

## Out of scope (v1)

- Light theme or system preference switching
- **Skeleton** loaders (optional follow-up; keep text “Loading…” unless added deliberately)
- New features or API changes
- Charting libraries (sparkline stays custom SVG)

## Testing and verification

- `web/`: `npm test`, `npm run lint`
- Playwright: `npm run test:e2e` (or repo-documented equivalent)
- **Preserve** e2e expectations where possible: headings (`Operator sign-in`, `Overview`, `Positions`, `Open`, `Closed (30d)`, section titles), links (`Optitrade Dashboard`, `Open positions →`, `← Overview`), **Sign in** button, login **alert**, **`data-testid="health-process-metrics"`**

## Follow-on

After this spec is accepted in the repo, create an implementation plan via the **writing-plans** workflow (tasks: init shadcn, wire aliases, define tokens, add `ui` components, migrate each file, run lint/test/e2e, fix regressions).
