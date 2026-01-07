import { test, expect } from '@playwright/test';

test.describe('Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should navigate to dashboard', async ({ page }) => {
    await page.getByRole('link', { name: /dashboard/i }).click();
    await expect(page).toHaveURL('/');
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
  });

  test('should navigate to videos list', async ({ page }) => {
    await page.getByRole('link', { name: /videos/i }).click();
    await expect(page).toHaveURL('/videos');
    await expect(page.getByRole('heading', { name: /all videos/i })).toBeVisible();
  });

  test('should navigate to search', async ({ page }) => {
    await page.getByRole('link', { name: /search/i }).click();
    await expect(page).toHaveURL('/search');
    await expect(page.getByRole('heading', { name: /search videos/i })).toBeVisible();
  });

  test('should navigate to cost analysis', async ({ page }) => {
    await page.getByRole('link', { name: /cost analysis/i }).click();
    await expect(page).toHaveURL('/costs');
    await expect(page.getByRole('heading', { name: /cost analysis/i })).toBeVisible();
  });

  test('should navigate to settings', async ({ page }) => {
    await page.getByRole('link', { name: /settings/i }).click();
    await expect(page).toHaveURL('/settings');
    await expect(page.getByRole('heading', { name: /settings/i })).toBeVisible();
  });

  test('should highlight active navigation item', async ({ page }) => {
    // Dashboard should be active
    const dashboardLink = page.getByRole('link', { name: /dashboard/i });
    await expect(dashboardLink).toHaveClass(/bg-primary|text-primary-foreground/);
    
    // Navigate to videos
    await page.getByRole('link', { name: /videos/i }).click();
    const videosLink = page.getByRole('link', { name: /videos/i });
    await expect(videosLink).toHaveClass(/bg-primary|text-primary-foreground/);
  });

  test('should show sidebar on all pages', async ({ page }) => {
    const pages = ['/', '/videos', '/search', '/costs', '/settings'];
    
    for (const path of pages) {
      await page.goto(path);
      await expect(page.getByText(/youtube analyzer/i)).toBeVisible();
      await expect(page.getByRole('navigation')).toBeVisible();
    }
  });

  test('should toggle dark mode', async ({ page }) => {
    const themeButton = page.getByRole('button', { name: /toggle theme/i }).or(
      page.locator('button').filter({ has: page.locator('svg') }).last()
    );
    
    await themeButton.click();
    
    // Should toggle dark class on html element
    const html = page.locator('html');
    const hasDark = await html.evaluate((el) => el.classList.contains('dark'));
    
    // Click again to toggle back
    await themeButton.click();
    const hasDarkAfter = await html.evaluate((el) => el.classList.contains('dark'));
    
    // Should have changed
    expect(hasDark).not.toBe(hasDarkAfter);
  });
});

