import { defineConfig, devices } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const webDir = path.dirname(fileURLToPath(import.meta.url))
const srcDir = path.join(webDir, '..', 'src')

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: 'list',
  use: {
    // Dedicated port so e2e does not attach to an unrelated dev server on 5173.
    baseURL: 'http://127.0.0.1:5180',
    trace: 'on-first-retry',
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: [
    {
      command: `bash -lc 'rm -f /tmp/optitrade-e2e.sqlite && env -u DERIBIT_CLIENT_ID -u DERIBIT_CLIENT_SECRET OPTITRADE_POLICY_PATH= OPTITRADE_SETTINGS_SECRET="${process.env.OPTITRADE_SETTINGS_SECRET ?? 'abcdefghijklmnopqrstuvwxyz123456'}" OPTITRADE_DASHBOARD_SESSION_PATH=/tmp/optitrade-e2e.sqlite go run -C "${srcDir}" ./cmd/optitrade dashboard -listen=127.0.0.1:8080'`,
      url: 'http://127.0.0.1:8080/healthz',
      timeout: 180_000,
      reuseExistingServer: !process.env.CI,
    },
    {
      command: 'npm run dev -- --host 127.0.0.1 --port 5180 --strictPort',
      cwd: webDir,
      url: 'http://127.0.0.1:5180',
      timeout: 120_000,
      reuseExistingServer: !process.env.CI,
    },
  ],
})
