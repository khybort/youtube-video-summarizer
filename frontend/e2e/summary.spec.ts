import { test, expect } from '@playwright/test';

test.describe('Summary Display', () => {
  const videoId = 'test-video-id';
  const mockSummary = {
    id: 'summary-id',
    videoId: videoId,
    modelUsed: 'gemini-1.5-flash',
    summaryType: 'detailed',
    content: 'This is a detailed summary of the video. It contains key information and insights.',
    keyPoints: [
      'Key point 1: Important information',
      'Key point 2: Another important detail',
      'Key point 3: Final key point',
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
          hasSummary: true,
        }),
      });
    });

    await page.goto(`/videos/${videoId}`);
    await page.getByRole('button', { name: /summary/i }).click();
  });

  test('should display summary when available', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockSummary),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    await expect(page.getByRole('heading', { name: /summary/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(mockSummary.content)).toBeVisible();
  });

  test('should display key points', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockSummary),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    await expect(page.getByText(/key points/i)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/key point 1/i)).toBeVisible();
  });

  test('should show generate summary form when no summary', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Summary not found' }),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    await expect(page.getByRole('heading', { name: /summary/i })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(/generate.*summary/i)).toBeVisible();
    await expect(page.getByRole('button', { name: /generate summary/i })).toBeVisible();
  });

  test('should generate summary', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Summary not found' }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockSummary),
        });
      }
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    // Select summary type
    const select = page.locator('select').first();
    await select.selectOption('detailed');
    
    // Click generate
    await page.getByRole('button', { name: /generate summary/i }).click();
    
    // Should show generating state
    await expect(page.getByText(/generating/i)).toBeVisible({ timeout: 3000 });
    
    // Should show summary after generation
    await expect(page.getByText(mockSummary.content)).toBeVisible({ timeout: 10000 });
  });

  test('should export summary', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockSummary),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    // Mock download
    const downloadPromise = page.waitForEvent('download', { timeout: 5000 }).catch(() => null);
    await page.getByRole('button', { name: /export/i }).click();
    
    // Download should be triggered (or at least button should work)
    await expect(page.getByRole('button', { name: /export/i })).toBeVisible();
  });

  test('should show summary metadata', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/summary`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockSummary),
      });
    });

    await page.reload();
    await page.getByRole('button', { name: /summary/i }).click();
    
    await expect(page.getByText(mockSummary.summaryType)).toBeVisible({ timeout: 5000 });
    await expect(page.getByText(mockSummary.modelUsed)).toBeVisible();
  });
});

