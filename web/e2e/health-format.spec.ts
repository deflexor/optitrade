import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('health panel formatting', () => {
  test('shows human-readable uptime and heap units', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/')
    await expect(page.getByRole('heading', { name: 'Overview' })).toBeVisible()

    const line = page.getByTestId('health-process-metrics')
    await expect(line).toBeVisible()
    const text = await line.innerText()
    expect(text.toLowerCase()).not.toContain('bytes')
    expect(text).toMatch(/uptime/i)
    expect(text).toMatch(/\d+\s*(s|m|h|d)/i)
    expect(text).toMatch(/\d+(\.\d+)?\s*(B|KiB|MiB|GiB)/)
  })
})
