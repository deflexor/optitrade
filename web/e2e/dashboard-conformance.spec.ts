import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('dashboard specification smoke', () => {
  test('overview: balance, P/L, mood, and strategy regions', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/')
    await expect(page.getByRole('heading', { name: 'Overview' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Balance' })).toBeVisible()
    await expect(
      page.getByRole('heading', { name: /P\/L/ }),
    ).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Market mood' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Strategy' })).toBeVisible()
  })

  test('navigation: overview to opportunities and back', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/')
    await page.getByRole('link', { name: 'Opportunities →' }).click()
    await expect(page).toHaveURL(/\/opportunities/)
    await expect(page.getByRole('heading', { name: 'Opportunities' })).toBeVisible()
    await page.getByRole('link', { name: 'Optitrade Dashboard' }).click()
    await expect(page).toHaveURL(/\/$/)
    await expect(page.getByRole('heading', { name: 'Overview' })).toBeVisible()
  })
})
