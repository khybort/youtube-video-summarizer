import { test, expect } from '@playwright/test';

test.describe('Search Page', () => {
  const mockVideos = [
    {
      id: '1',
      title: 'React Tutorial',
      channelName: 'Tech Channel',
      viewCount: 1000,
      status: 'completed',
    },
    {
      id: '2',
      title: 'Vue.js Guide',
      channelName: 'Dev Channel',
      viewCount: 2000,
      status: 'processing',
    },
  ];

  test.beforeEach(async ({ page }) => {
    await page.goto('/search');
  });

  test('should display search page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /search videos/i })).toBeVisible();
    await expect(page.getByText(/find videos by title/i)).toBeVisible();
  });

  test('should show search input', async ({ page }) => {
    await expect(page.getByPlaceholder(/search videos/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /search/i })).toBeVisible();
  });

  test('should perform search', async ({ page }) => {
    await page.route('**/api/v1/videos*', async (route) => {
      const url = new URL(route.request().url());
      const search = url.searchParams.get('search');
      
      if (search === 'react') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            videos: [mockVideos[0]],
            total: 1,
          }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            videos: [],
            total: 0,
          }),
        });
      }
    });

    await page.getByPlaceholder(/search videos/i).fill('react');
    await page.getByRole('button', { name: /search/i }).click();
    
    await expect(page.getByText(/react tutorial/i)).toBeVisible({ timeout: 5000 });
  });

  test('should show search results', async ({ page }) => {
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: mockVideos,
          total: 2,
        }),
      });
    });

    await page.getByPlaceholder(/search videos/i).fill('test');
    await page.getByRole('button', { name: /search/i }).click();
    
    await expect(page.getByText(/found 2 video/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/react tutorial/i)).toBeVisible();
    await expect(page.getByText(/vue\.js guide/i)).toBeVisible();
  });

  test('should filter by status', async ({ page }) => {
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [mockVideos[0]], // Only completed
          total: 1,
        }),
      });
    });

    const statusSelect = page.locator('select').first();
    await statusSelect.selectOption('completed');
    
    // Should trigger search or filter
    await expect(page.getByText(/react tutorial/i)).toBeVisible({ timeout: 5000 });
  });

  test('should show empty state when no results', async ({ page }) => {
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [],
          total: 0,
        }),
      });
    });

    await page.getByPlaceholder(/search videos/i).fill('nonexistent');
    await page.getByRole('button', { name: /search/i }).click();
    
    await expect(page.getByText(/no videos found/i)).toBeVisible({ timeout: 5000 });
  });

  test('should clear filters', async ({ page }) => {
    await page.getByPlaceholder(/search videos/i).fill('test');
    await page.getByRole('button', { name: /search/i }).click();
    
    // Clear button should appear
    await expect(page.getByRole('button', { name: /clear/i })).toBeVisible();
    
    await page.getByRole('button', { name: /clear/i }).click();
    
    // Search input should be cleared
    await expect(page.getByPlaceholder(/search videos/i)).toHaveValue('');
  });

  test('should show filters section', async ({ page }) => {
    await expect(page.getByText(/filters/i)).toBeVisible();
    await expect(page.getByText(/all status/i)).toBeVisible();
  });
});

