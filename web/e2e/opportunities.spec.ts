import { expect, test } from '@playwright/test'
import { signInAsDevOperator } from './helpers/login'

test.describe('opportunities', () => {
  test('shows heading and table region', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/opportunities')
    await expect(page.getByRole('heading', { name: 'Opportunities' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Status' })).toBeVisible()
  })

  test('paused mode shows banner', async ({ page }) => {
    await signInAsDevOperator(page)
    await page.goto('/settings')
    await page.getByLabel('Client ID').fill('e2e_client_id')
    await page.getByLabel('Client secret').fill('e2e_client_secret')
    await page.getByRole('button', { name: 'Save settings' }).click()
    await expect(page.getByRole('button', { name: 'Save settings' })).toBeEnabled({
      timeout: 15_000,
    })
    await page.goto('/opportunities')
    await expect(page.getByRole('heading', { name: 'Opportunities' })).toBeVisible()
    await page.getByRole('button', { name: 'Paused' }).click()
    await expect(page.getByRole('alert')).toBeVisible({ timeout: 15_000 })
    await expect(page.getByText(/paused/i).first()).toBeVisible()
  })
})
