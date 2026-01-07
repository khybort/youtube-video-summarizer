import { test, expect } from '@playwright/test';

/**
 * Integration test: Complete user flow
 * Tests the full workflow from adding a video to viewing summary
 */
test.describe('Complete User Flow', () => {
  test('should complete full video analysis workflow', async ({ page }) => {
    const videoUrl = 'https://www.youtube.com/watch?v=dQw4w9WgXcQ';
    const videoId = 'test-video-id';

    // Step 1: Navigate to dashboard
    await page.goto('/');
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();

    // Step 2: Add a video
    await page.route('**/api/v1/videos', async (route) => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            id: videoId,
            youtubeId: 'dQw4w9WgXcQ',
            title: 'Test Video',
            channelName: 'Test Channel',
            status: 'pending',
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

    await page.getByPlaceholder(/https:\/\/www\.youtube\.com\/watch\?v=/i).fill(videoUrl);
    await page.getByRole('button', { name: /add video/i }).click();

    // Step 3: Navigate to videos list
    await page.getByRole('link', { name: /videos/i }).click();
    await expect(page).toHaveURL('/videos');

    // Step 4: Mock video detail and navigate
    await page.route(`**/api/v1/videos/${videoId}`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: videoId,
          title: 'Test Video',
          channelName: 'Test Channel',
          status: 'completed',
          hasTranscript: true,
          hasSummary: true,
        }),
      });
    });

    await page.getByText(/test video/i).first().click();
    await expect(page).toHaveURL(`/videos/${videoId}`);

    // Step 5: View transcript
    await page.route(`**/api/v1/videos/${videoId}/transcript`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'transcript-id',
          videoId: videoId,
          language: 'en',
          source: 'whisper',
          content: 'Test transcript content',
          segments: [{ start: 0, end: 10, text: 'Test transcript content' }],
        }),
      });
    });

    await page.getByRole('button', { name: /transcript/i }).click();
    await expect(page.getByRole('heading', { name: /transcript/i })).toBeVisible({ timeout: 5000 });

    // Step 6: View summary
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: 'summary-id',
          videoId: videoId,
          modelUsed: 'gemini-1.5-flash',
          summaryType: 'detailed',
          content: 'Test summary content',
          keyPoints: ['Point 1', 'Point 2'],
        }),
      });
    });

    await page.getByRole('button', { name: /summary/i }).click();
    await expect(page.getByRole('heading', { name: /summary/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/test summary content/i)).toBeVisible();

    // Step 7: View similar videos
    await page.route(`**/api/v1/videos/${videoId}/similar*`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          similar_videos: [
            {
              video: {
                id: '2',
                title: 'Similar Video',
                similarityScore: 0.85,
              },
              similarityScore: 0.85,
            },
          ],
        }),
      });
    });

    await page.getByRole('button', { name: /similar/i }).click();
    await expect(page.getByText(/similar videos/i)).toBeVisible({ timeout: 5000 });
  });

  test('should handle error states gracefully', async ({ page }) => {
    // Test error handling in various scenarios
    await page.goto('/');

    // Mock API error
    await page.route('**/api/v1/videos', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' }),
      });
    });

    await page.getByPlaceholder(/https:\/\/www\.youtube\.com\/watch\?v=/i).fill('https://www.youtube.com/watch?v=test');
    await page.getByRole('button', { name: /add video/i }).click();

    // Should show error message
    await expect(page.getByText(/error/i)).toBeVisible({ timeout: 5000 });
  });
});

