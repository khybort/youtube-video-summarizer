import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display dashboard page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
    await expect(page.getByText(/add and analyze youtube videos/i)).toBeVisible();
  });

  test('should show add video form', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /add new video/i })).toBeVisible();
    await expect(page.getByPlaceholder(/https:\/\/www\.youtube\.com\/watch\?v=/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /add video/i })).toBeVisible();
  });

  test('should add a video successfully', async ({ page }) => {
    // Mock API response
    await page.route('**/api/v1/videos', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: 'test-video-id',
            youtubeId: 'dQw4w9WgXcQ',
            title: 'Test Video',
            channelName: 'Test Channel',
            status: 'pending',
          }),
        });
      } else {
        await route.continue();
      }
    });

    const videoUrl = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
    
    await page.getByPlaceholder(/https:\/\/www\.youtube\.com\/watch\?v=/i).fill(videoUrl);
    await page.getByRole('button', { name: /add video/i }).click();

    // Wait for success message or video to appear
    await expect(page.getByText(/test video/i).first()).toBeVisible({ timeout: 10000 });
  });

  test('should show error for invalid URL', async ({ page }) => {
    await page.route('**/api/v1/videos', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Invalid YouTube URL' }),
        });
      } else {
        await route.continue();
      }
    });

    await page.getByPlaceholder(/https:\/\/www\.youtube\.com\/watch\?v=/i).fill('invalid-url');
    await page.getByRole('button', { name: /add video/i }).click();

    // Should show error message
    await expect(page.getByText(/error|invalid/i)).toBeVisible({ timeout: 5000 });
  });

  test('should display recent videos', async ({ page }) => {
    // Mock API response
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [
            {
              id: '1',
              title: 'Test Video 1',
              channelName: 'Channel 1',
              viewCount: 1000,
              status: 'completed',
            },
            {
              id: '2',
              title: 'Test Video 2',
              channelName: 'Channel 2',
              viewCount: 2000,
              status: 'processing',
            },
          ],
          total: 2,
        }),
      });
    });

    await page.reload();
    
    await expect(page.getByRole('heading', { name: /recent videos/i })).toBeVisible();
    await expect(page.getByText(/test video 1/i)).toBeVisible();
    await expect(page.getByText(/test video 2/i)).toBeVisible();
  });

  test('should show empty state when no videos', async ({ page }) => {
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

    await page.reload();
    
    await expect(page.getByText(/no videos yet/i)).toBeVisible();
  });
});

