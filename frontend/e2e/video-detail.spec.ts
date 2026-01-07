import { test, expect } from '@playwright/test';

test.describe('Video Detail', () => {
  const videoId = 'test-video-id';
  const mockVideo = {
    id: videoId,
    youtubeId: 'dQw4w9WgXcQ',
    title: 'Test Video Title',
    description: 'This is a test video description',
    channelName: 'Test Channel',
    channelId: 'channel123',
    viewCount: 10000,
    likeCount: 500,
    duration: 300,
    status: 'completed',
    thumbnailUrl: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg',
    hasTranscript: true,
    hasSummary: true,
    publishedAt: '2024-01-01T00:00:00Z',
  };

  test.beforeEach(async ({ page }) => {
    // Mock video API
    await page.route(`**/api/v1/videos/${videoId}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockVideo),
      });
    });

    await page.goto(`/videos/${videoId}`);
  });

  test('should display video detail page', async ({ page }) => {
    await expect(page.getByText(mockVideo.title)).toBeVisible();
    await expect(page.getByText(mockVideo.channelName)).toBeVisible();
  });

  test('should display video player', async ({ page }) => {
    // React Player should be rendered
    const player = page.locator('[data-testid="react-player"], iframe[src*="youtube"]').first();
    await expect(player).toBeVisible({ timeout: 10000 });
  });

  test('should show video statistics', async ({ page }) => {
    await expect(page.getByText(/10K views/i)).toBeVisible();
    await expect(page.getByText(/500 likes/i)).toBeVisible();
    await expect(page.getByText(/5:00/i)).toBeVisible(); // duration
  });

  test('should display video description', async ({ page }) => {
    await expect(page.getByText(mockVideo.description)).toBeVisible();
  });

  test('should show video info sidebar', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /video info/i })).toBeVisible();
    await expect(page.getByText(/status/i)).toBeVisible();
    await expect(page.getByText(/published/i)).toBeVisible();
  });

  test('should display cost breakdown card', async ({ page }) => {
    // Mock cost usage API
    await page.route(`**/api/v1/costs/videos/${videoId}/usage`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          usage: [
            {
              id: '1',
              videoId: videoId,
              operation: 'summarization',
              provider: 'gemini',
              model: 'gemini-1.5-flash',
              inputTokens: 1000,
              outputTokens: 500,
              totalTokens: 1500,
              cost: 0.0015,
              createdAt: new Date().toISOString(),
            },
          ],
        }),
      });
    });

    await page.reload();

    // Cost breakdown should be visible
    await expect(page.getByText(/cost breakdown/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/total cost/i)).toBeVisible();
  });

  test('should navigate between tabs', async ({ page }) => {
    // Overview tab (default)
    await expect(page.getByText(/description/i)).toBeVisible();

    // Transcript tab
    await page.getByRole('button', { name: /transcript/i }).click();
    await expect(page.getByRole('heading', { name: /transcript/i })).toBeVisible({ timeout: 5000 });

    // Summary tab
    await page.getByRole('button', { name: /summary/i }).click();
    await expect(page.getByRole('heading', { name: /summary/i })).toBeVisible({ timeout: 5000 });

    // Similar videos tab
    await page.getByRole('button', { name: /similar/i }).click();
    await expect(page.getByText(/similar videos/i)).toBeVisible({ timeout: 5000 });
  });

  test('should start video analysis', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/analyze`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Analysis started', status: 'processing' }),
      });
    });

    await page.getByRole('button', { name: /analyze video/i }).click();
    
    // Should show processing state
    await expect(page.getByText(/analyzing/i)).toBeVisible({ timeout: 5000 });
  });

  test('should navigate back to videos list', async ({ page }) => {
    await page.getByRole('button', { name: /back/i }).click();
    await expect(page).toHaveURL(/\/videos/);
  });
});

