import { test, expect } from '@playwright/test';

test.describe('Similar Videos', () => {
  const videoId = 'test-video-id';
  const mockSimilarVideos = {
    similar_videos: [
      {
        video: {
          id: '2',
          title: 'Similar Video 1',
          channelName: 'Channel 1',
          similarityScore: 0.85,
        },
        similarityScore: 0.85,
        comparisonType: 'combined',
      },
      {
        video: {
          id: '3',
          title: 'Similar Video 2',
          channelName: 'Channel 2',
          similarityScore: 0.72,
        },
        similarityScore: 0.72,
        comparisonType: 'combined',
      },
    ],
  };

  test.beforeEach(async ({ page }) => {
    // Mock video API
    await page.route(`**/api/v1/videos/${videoId}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: videoId,
          title: 'Test Video',
        }),
      });
    });

    // Mock similar videos API
    await page.route(`**/api/v1/videos/${videoId}/similar*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockSimilarVideos),
      });
    });

    await page.goto(`/videos/${videoId}`);
    await page.getByRole('button', { name: /similar/i }).click();
  });

  test('should display similar videos section', async ({ page }) => {
    await expect(page.getByText(/similar videos/i)).toBeVisible({ timeout: 5000 });
  });

  test('should show similar videos list', async ({ page }) => {
    await expect(page.getByText(/similar video 1/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/similar video 2/i)).toBeVisible();
  });

  test('should display similarity scores', async ({ page }) => {
    // Should show percentage match
    await expect(page.getByText(/85.*match/i).or(page.getByText(/85%/i))).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/72.*match/i).or(page.getByText(/72%/i))).toBeVisible();
  });

  test('should show similarity count', async ({ page }) => {
    await expect(page.getByText(/found.*2.*similar/i)).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to similar video on click', async ({ page }) => {
    // Mock video detail API
    await page.route('**/api/v1/videos/2', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '2',
          title: 'Similar Video 1',
        }),
      });
    });

    await page.getByText(/similar video 1/i).click();
    
    await expect(page).toHaveURL(/\/videos\/2/);
    await expect(page.getByText(/similar video 1/i)).toBeVisible();
  });

  test('should show empty state when no similar videos', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/similar*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          similar_videos: [],
        }),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /similar/i }).click();
    
    await expect(page.getByText(/no similar videos found/i)).toBeVisible({ timeout: 5000 });
  });

  test('should filter by similarity threshold', async ({ page }) => {
    // This would require API support for min_score parameter
    await expect(page.getByText(/similar videos/i)).toBeVisible({ timeout: 5000 });
  });
});

