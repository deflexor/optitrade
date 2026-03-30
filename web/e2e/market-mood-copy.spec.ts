import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('market mood copy', () => {
  test('does not show internal placeholder wording', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/')
    await expect(page.getByRole('heading', { name: 'Market mood' })).toBeVisible()
    await expect(page.getByText('strategy modules not wired')).toHaveCount(0)
    await expect(page.getByText(/not wired to dashboard/i)).toHaveCount(0)
  })
})
