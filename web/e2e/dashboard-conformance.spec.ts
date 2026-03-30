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

  test('navigation: overview to positions and back', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/')
    await page.getByRole('link', { name: 'Open positions →' }).click()
    await expect(page).toHaveURL(/\/positions/)
    await expect(page.getByRole('heading', { name: 'Positions' })).toBeVisible()
    await page.getByRole('link', { name: '← Overview' }).click()
    await expect(page).toHaveURL(/\/$/)
    await expect(page.getByRole('heading', { name: 'Overview' })).toBeVisible()
  })

  test('positions page lists open and closed sections', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/positions')
    await expect(page.getByRole('heading', { name: 'Open' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Closed (30d)' })).toBeVisible()
  })
})
