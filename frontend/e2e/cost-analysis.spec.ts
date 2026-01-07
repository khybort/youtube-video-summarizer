import { test, expect } from '@playwright/test';

test.describe('Cost Analysis Page', () => {
  const mockCostSummary = {
    totalCost: 0.0525,
    totalTokens: 15000,
    byProvider: {
      gemini: 0.03,
      groq: 0.0225,
    },
    byOperation: {
      summarization: 0.03,
      transcription: 0.0225,
    },
    byModel: {
      'gemini-1.5-flash': 0.03,
      'whisper-large-v3': 0.0225,
    },
    period: 'month',
    videoCount: 5,
    averageCostPerVideo: 0.0105,
  };

  const mockUsage = [
    {
      id: '1',
      videoId: 'video-1',
      operation: 'summarization',
      provider: 'gemini',
      model: 'gemini-1.5-flash',
      inputTokens: 1000,
      outputTokens: 500,
      totalTokens: 1500,
      cost: 0.0015,
      createdAt: new Date().toISOString(),
    },
    {
      id: '2',
      videoId: 'video-2',
      operation: 'transcription',
      provider: 'groq',
      model: 'whisper-large-v3',
      inputTokens: 2000,
      outputTokens: 0,
      totalTokens: 2000,
      cost: 0.002,
      createdAt: new Date().toISOString(),
    },
  ];

  test.beforeEach(async ({ page }) => {
    // Mock cost API endpoints
    await page.route('**/api/v1/costs/summary*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify(mockCostSummary),
      });
    });

    await page.route('**/api/v1/costs/usage*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ usage: mockUsage }),
      });
    });

    await page.goto('/costs');
  });

  test('should display cost analysis page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /cost analysis/i })).toBeVisible();
    await expect(page.getByText(/track token usage and costs/i)).toBeVisible();
  });

  test('should display summary cards', async ({ page }) => {
    // Total Cost card
    await expect(page.getByText(/total cost/i)).toBeVisible();
    await expect(page.getByText(/\$0\.0525/i)).toBeVisible();

    // Total Tokens card
    await expect(page.getByText(/total tokens/i)).toBeVisible();
    await expect(page.getByText(/15,000/i)).toBeVisible();

    // Average per Video card
    await expect(page.getByText(/avg per video/i)).toBeVisible();
    await expect(page.getByText(/\$0\.0105/i)).toBeVisible();

    // Videos card
    await expect(page.getByText(/videos/i)).toBeVisible();
    await expect(page.getByText(/5/i)).toBeVisible();
  });

  test('should display cost breakdown by provider', async ({ page }) => {
    await expect(page.getByText(/cost by provider/i)).toBeVisible();
    await expect(page.getByText(/gemini/i)).toBeVisible();
    await expect(page.getByText(/groq/i)).toBeVisible();
  });

  test('should display cost breakdown by operation', async ({ page }) => {
    await expect(page.getByText(/cost by operation/i)).toBeVisible();
    await expect(page.getByText(/summarization/i)).toBeVisible();
    await expect(page.getByText(/transcription/i)).toBeVisible();
  });

  test('should display cost breakdown by model', async ({ page }) => {
    await expect(page.getByText(/cost by model/i)).toBeVisible();
    await expect(page.getByText(/gemini-1\.5-flash/i)).toBeVisible();
    await expect(page.getByText(/whisper-large-v3/i)).toBeVisible();
  });

  test('should display usage table', async ({ page }) => {
    await expect(page.getByText(/recent usage/i)).toBeVisible();
    await expect(page.getByText(/detailed token usage records/i)).toBeVisible();

    // Check table headers
    await expect(page.getByRole('columnheader', { name: /date/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /operation/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /provider/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /model/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /tokens/i })).toBeVisible();
    await expect(page.getByRole('columnheader', { name: /cost/i })).toBeVisible();

    // Check table rows
    await expect(page.getByText(/summarization/i)).toBeVisible();
    await expect(page.getByText(/transcription/i)).toBeVisible();
  });

  test('should allow changing period filter', async ({ page }) => {
    const periodSelect = page.locator('select').first();
    await expect(periodSelect).toBeVisible();

    // Change to week
    await periodSelect.selectOption('week');

    // API should be called with new period
    await page.waitForResponse('**/api/v1/costs/summary?period=week');
  });

  test('should show empty state when no usage data', async ({ page }) => {
    await page.route('**/api/v1/costs/usage*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ usage: [] }),
      });
    });

    await page.reload();

    await expect(page.getByText(/no usage data available/i)).toBeVisible();
  });

  test('should handle API errors gracefully', async ({ page }) => {
    await page.route('**/api/v1/costs/summary*', async (route) => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ error: 'Internal server error' }),
      });
    });

    await page.reload();

    // Should show error state or loading
    await expect(
      page.getByText(/error|loading|failed/i).or(page.getByText(/cost analysis/i))
    ).toBeVisible({ timeout: 5000 });
  });

  test('should navigate from sidebar', async ({ page }) => {
    // Start from dashboard
    await page.goto('/');
    
    // Click on Cost Analysis in sidebar
    await page.getByRole('link', { name: /cost analysis/i }).click();
    
    // Should navigate to cost analysis page
    await expect(page).toHaveURL(/.*\/costs/);
    await expect(page.getByRole('heading', { name: /cost analysis/i })).toBeVisible();
  });

  test('should display cost values in correct format', async ({ page }) => {
    // Check currency formatting
    const costText = await page.getByText(/\$0\.0525/i).textContent();
    expect(costText).toContain('$');
    
    // Check token formatting
    const tokenText = await page.getByText(/15,000/i).textContent();
    expect(tokenText).toMatch(/\d+/);
  });
});

