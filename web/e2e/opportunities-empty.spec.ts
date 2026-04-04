import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('opportunities without venue policy', () => {
  test('page loads and shows grid or empty copy', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/opportunities')
    await expect(page.getByRole('heading', { name: 'Opportunities' })).toBeVisible()
    await expect(page.getByRole('table')).toBeVisible()
  })
})
