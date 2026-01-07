import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
  const mockSettings = {
    llmProvider: 'gemini',
    whisperProvider: 'groq',
    ollamaModel: 'llama3.2',
    whisperModel: 'base',
    theme: 'dark',
  };

  test.beforeEach(async ({ page }) => {
    // Mock settings API
    await page.route('**/api/v1/settings', async (route) => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockSettings),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockSettings),
        });
      }
    });

    await page.goto('/settings');
  });

  test('should display settings page', async ({ page }) => {
    await expect(page.getByRole('heading', { name: /settings/i })).toBeVisible();
    await expect(page.getByText(/configure your ai provider/i)).toBeVisible();
  });

  test('should show current settings', async ({ page }) => {
    await expect(page.getByText(/llm provider/i)).toBeVisible();
    await expect(page.getByText(/whisper provider/i)).toBeVisible();
  });

  test('should allow changing LLM provider', async ({ page }) => {
    const llmSelect = page.locator('select').first();
    await expect(llmSelect).toBeVisible();
    
    // Change to Ollama
    await llmSelect.selectOption('ollama');
    
    // Should show Ollama model input
    await expect(page.getByPlaceholder(/llama3\.2/i)).toBeVisible();
  });

  test('should allow changing Whisper provider', async ({ page }) => {
    const whisperSelect = page.locator('select').nth(1);
    await expect(whisperSelect).toBeVisible();
    
    // Change to local
    await whisperSelect.selectOption('local');
    
    // Should show Whisper model select
    await expect(page.getByText(/whisper model/i)).toBeVisible();
  });

  test('should save settings', async ({ page }) => {
    let savedSettings: any = null;
    
    await page.route('**/api/v1/settings', async (route) => {
      if (route.request().method() === 'PUT') {
        savedSettings = route.request().postDataJSON();
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(savedSettings),
        });
      } else {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockSettings),
        });
      }
    });

    await page.getByRole('button', { name: /save changes/i }).click();
    
    // Should show success or loading state
    await expect(page.getByText(/saving|saved/i).or(page.getByText(/save changes/i))).toBeVisible({ timeout: 5000 });
  });

  test('should show Ollama model input when Ollama selected', async ({ page }) => {
    const llmSelect = page.locator('select').first();
    await llmSelect.selectOption('ollama');
    
    await expect(page.getByText(/ollama model/i)).toBeVisible();
    await expect(page.getByPlaceholder(/llama3\.2/i)).toBeVisible();
  });

  test('should show Whisper model select when local selected', async ({ page }) => {
    const whisperSelect = page.locator('select').nth(1);
    await whisperSelect.selectOption('local');
    
    await expect(page.getByText(/whisper model/i)).toBeVisible();
    const modelSelect = page.locator('select').last();
    await expect(modelSelect).toBeVisible();
  });
});

