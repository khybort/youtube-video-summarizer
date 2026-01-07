import { test, expect } from '@playwright/test';

test.describe('Video Actions', () => {
  const videoId = 'test-video-id';
  const mockVideo = {
    id: videoId,
    title: 'Test Video',
    channelName: 'Test Channel',
    status: 'pending',
  };

  test.beforeEach(async ({ page }) => {
    // Mock videos API
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [mockVideo],
          total: 1,
        }),
      });
    });

    await page.goto('/videos');
  });

  test('should show video actions menu', async ({ page }) => {
    // Find the more options button (three dots)
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.hover();
    
    // Should show dropdown menu
    await expect(page.getByText(/analyze/i).or(page.getByText(/delete/i))).toBeVisible({ timeout: 3000 });
  });

  test('should start video analysis', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}/analyze`, async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Analysis started' }),
      });
    });

    // Open actions menu
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.click();
    
    // Click analyze
    await page.getByText(/analyze/i).click();
    
    // Should show success or processing state
    await expect(page.getByText(/analyzing|analysis started/i)).toBeVisible({ timeout: 5000 });
  });

  test('should show delete confirmation dialog', async ({ page }) => {
    // Open actions menu
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.click();
    
    // Click delete
    await page.getByText(/delete/i).click();
    
    // Should show confirmation dialog
    await expect(page.getByRole('heading', { name: /delete video/i })).toBeVisible({ timeout: 3000 });
    await expect(page.getByText(/are you sure/i)).toBeVisible();
  });

  test('should delete video on confirmation', async ({ page }) => {
    await page.route(`**/api/v1/videos/${videoId}`, async (route) => {
      if (route.request().method() === 'DELETE') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ message: 'Video deleted' }),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockVideo),
        });
      }
    });

    // Open actions menu
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.click();
    
    // Click delete
    await page.getByText(/delete/i).click();
    
    // Confirm deletion
    await page.getByRole('button', { name: /delete$/i }).click();
    
    // Video should be removed or show success message
    await expect(page.getByText(/deleted|success/i)).toBeVisible({ timeout: 5000 });
  });

  test('should cancel delete action', async ({ page }) => {
    // Open actions menu
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.click();
    
    // Click delete
    await page.getByText(/delete/i).click();
    
    // Cancel deletion
    await page.getByRole('button', { name: /cancel/i }).click();
    
    // Dialog should close
    await expect(page.getByRole('heading', { name: /delete video/i })).not.toBeVisible();
    
    // Video should still be visible
    await expect(page.getByText(mockVideo.title)).toBeVisible();
  });

  test('should disable analyze button when processing', async ({ page }) => {
    const processingVideo = { ...mockVideo, status: 'processing' };
    
    await page.route('**/api/v1/videos*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          videos: [processingVideo],
          total: 1,
        }),
      });
    });

    await page.reload();
    
    // Open actions menu
    const actionsButton = page.locator('button').filter({ has: page.locator('svg') }).first();
    await actionsButton.click();
    
    // Analyze button should be disabled
    const analyzeButton = page.getByText(/analyze/i);
    await expect(analyzeButton).toBeDisabled();
  });
});

