import { test, expect } from '@playwright/test';

test.describe('Video List', () => {
  test.beforeEach(async ({ page }) => {
    // Mock videos API
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [
            {
              id: '1',
              youtubeId: 'dQw4w9WgXcQ',
              title: 'Test Video 1',
              channelName: 'Test Channel 1',
              viewCount: 1000,
              likeCount: 100,
              duration: 180,
              status: 'completed',
              thumbnailUrl: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg',
              hasTranscript: true,
              hasSummary: true,
            },
            {
              id: '2',
              youtubeId: 'test2',
              title: 'Test Video 2',
              channelName: 'Test Channel 2',
              viewCount: 2000,
              likeCount: 200,
              duration: 240,
              status: 'processing',
              thumbnailUrl: 'https://i.ytimg.com/vi/test2/default.jpg',
              hasTranscript: false,
              hasSummary: false,
            },
          ],
          total: 2,
        }),
      });
    });

    await page.goto('/videos');
  });

  test('should display video list page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /all videos/i })).toBeVisible();
    await expect(page.getByText(/2 video/i)).toBeVisible();
  });

  test('should display video cards', async ({ page }) => {
    await expect(page.getByText(/test video 1/i)).toBeVisible();
    await expect(page.getByText(/test video 2/i)).toBeVisible();
    await expect(page.getByText(/test channel 1/i)).toBeVisible();
    await expect(page.getByText(/test channel 2/i)).toBeVisible();
  });

  test('should navigate to video detail on click', async ({ page }) => {
    // Mock video detail API
    await page.route('**/api/v1/videos/1', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1',
          title: 'Test Video 1',
          channelName: 'Test Channel 1',
        }),
      });
    });

    await page.getByText(/test video 1/i).first().click();
    
    await expect(page).toHaveURL(/\/videos\/1/);
    await expect(page.getByText(/test video 1/i)).toBeVisible();
  });

  test('should show video thumbnails', async ({ page }) => {
    const thumbnails = page.locator('img[src*="ytimg.com"]');
    await expect(thumbnails.first()).toBeVisible();
  });

  test('should show video status badges', async ({ page }) => {
    await expect(page.getByText(/completed/i)).toBeVisible();
    await expect(page.getByText(/processing/i)).toBeVisible();
  });

  test('should show view counts', async ({ page }) => {
    await expect(page.getByText(/1K views/i)).toBeVisible();
    await expect(page.getByText(/2K views/i)).toBeVisible();
  });
});

