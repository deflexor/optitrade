import { expect, type Page } from '@playwright/test'

/** Default Playwright / dev operator (see dashboard allowlist and playwright webServer). */
export async function signInAsDevOperator(page: Page) {
  await page.goto('/login')
  await expect(page.getByRole('heading', { name: 'Operator sign-in' })).toBeVisible()
  await page.getByLabel('Username').fill('opti')
  await page.getByLabel('Password').fill('opti')
  await page.getByRole('button', { name: 'Sign in' }).click()
  await expect(page.getByRole('link', { name: 'Optitrade Dashboard' })).toBeVisible()
}
