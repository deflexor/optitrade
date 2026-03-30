import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('positions without exchange', () => {
  test('shows unavailable state for open and closed lists', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/positions')
    await expect(page.getByRole('heading', { name: 'Positions' })).toBeVisible()

    const openSection = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Open' }) })
    const closedSection = page.locator('section').filter({ has: page.getByRole('heading', { name: 'Closed (30d)' }) })

    await expect(openSection.getByText('Loading…')).toHaveCount(0)
    await expect(closedSection.getByText('Loading…')).toHaveCount(0)

    await expect(openSection.getByText(/unavailable|configured/i)).toBeVisible()
    await expect(closedSection.getByText(/unavailable|configured/i)).toBeVisible()
  })
})
