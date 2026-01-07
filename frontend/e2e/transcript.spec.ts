import { test, expect } from '@playwright/test';

test.describe('Transcript Viewer', () => {
  const videoId = 'test-video-id';
  const mockTranscript = {
    id: 'transcript-id',
    videoId: videoId,
    language: 'en',
    source: 'whisper',
    content: 'This is a full transcript text. It contains multiple sentences.',
    segments: [
      { start: 0, end: 5, text: 'This is a full transcript text.' },
      { start: 5, end: 10, text: 'It contains multiple sentences.' },
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
          hasTranscript: true,
        }),
      });
    });

    // Mock transcript API
    await page.route(`**/api/v1/videos/${videoId}/transcript`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockTranscript),
      });
    });

    await page.goto(`/videos/${videoId}`);
    await page.getByRole('button', { name: /transcript/i }).click();
  });

  test('should display transcript viewer', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /transcript/i })).toBeVisible();
    await expect(page.getByText(/en.*whisper/i)).toBeVisible();
  });

  test('should display transcript segments', async ({ page }) => {
    await expect(page.getByText(mockTranscript.segments[0].text)).toBeVisible();
    await expect(page.getByText(mockTranscript.segments[1].text)).toBeVisible();
  });

  test('should show timestamps', async ({ page }) => {
    await expect(page.getByText(/0s/i)).toBeVisible();
    await expect(page.getByText(/5s/i)).toBeVisible();
  });

  test('should search transcript', async ({ page }) => {
    const searchInput = page.getByPlaceholder(/search transcript/i);
    await searchInput.fill('multiple');
    
    // Should filter segments
    await expect(page.getByText(/multiple sentences/i)).toBeVisible();
    await expect(page.getByText(/full transcript text/i)).not.toBeVisible();
  });

  test('should copy transcript to clipboard', async ({ page }) => {
    // Mock clipboard API
    await page.context().grantPermissions(['clipboard-read', 'clipboard-write']);

    await page.getByRole('button', { name: /copy/i }).click();
    
    // Should show copied confirmation
    await expect(page.getByText(/copied/i)).toBeVisible({ timeout: 3000 });
  });

  test('should highlight active segment on click', async ({ page }) => {
    const firstSegment = page.getByText(mockTranscript.segments[0].text).first();
    await firstSegment.click();
    
    // Segment should be highlighted (check for active class or style)
    await expect(firstSegment).toHaveClass(/active|bg-primary/, { timeout: 1000 });
  });

  test('should show empty state when no transcript', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/transcript`, async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Transcript not found' }),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /transcript/i }).click();
    
    await expect(page.getByText(/no transcript available/i)).toBeVisible({ timeout: 5000 });
  });
});

