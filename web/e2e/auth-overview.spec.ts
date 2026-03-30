import { expect, test } from '@playwright/test'

test.describe('operator dashboard', () => {
  test('health API is public', async ({ request }) => {
    const res = await request.get('http://127.0.0.1:8080/api/v1/health')
    expect(res.ok()).toBeTruthy()
    const body = (await res.json()) as { trading?: { exchange_reachable?: boolean; detail?: string } }
    expect(body.trading?.exchange_reachable).toBe(false)
    expect(body.trading?.detail).toBe('exchange_not_configured')
  })

  test('login with default dev operator and see overview shell', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByRole('heading', { name: 'Operator sign-in' })).toBeVisible()

    await page.getByLabel('Username').fill('opti')
    await page.getByLabel('Password').fill('opti')
    await page.getByRole('button', { name: 'Sign in' }).click()

    await expect(page.getByRole('link', { name: 'Optitrade Dashboard' })).toBeVisible()
    await expect(page.locator('header').getByText('opti', { exact: true })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Overview' })).toBeVisible()
    await expect(
      page.getByText(/Exchange data partially degraded/),
    ).toBeVisible()
  })

  test('rejects wrong password', async ({ page }) => {
    await page.goto('/login')
    await page.getByLabel('Username').fill('opti')
    await page.getByLabel('Password').fill('wrong-password')
    await page.getByRole('button', { name: 'Sign in' }).click()
    await expect(page.getByRole('alert')).toContainText('Sign-in failed')
  })
})
