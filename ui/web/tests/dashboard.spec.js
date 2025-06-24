import { test, expect } from '@playwright/test';

test('dashboard graphs load', async ({ page }) => {
  await page.goto('http://localhost:8080');
  await expect(page.getByText('Running Agents')).toBeVisible();
  await expect(page.locator('#tokChart')).toBeVisible();
  await expect(page.locator('#usageChart')).toBeVisible();
  await expect(page.locator('#healthChart')).toBeVisible();
});
