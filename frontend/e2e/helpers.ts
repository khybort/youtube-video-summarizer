import { Page } from '@playwright/test';

/**
 * Helper function to wait for API response
 */
export async function waitForAPI(page: Page, urlPattern: string | RegExp, timeout = 10000) {
  return page.waitForResponse(
    (response) => {
      const url = response.url();
      if (typeof urlPattern === 'string') {
        return url.includes(urlPattern);
      }
      return urlPattern.test(url);
    },
    { timeout }
  );
}

/**
 * Helper function to mock API response
 */
export async function mockAPI(
  page: Page,
  urlPattern: string | RegExp,
  response: any,
  status = 200
) {
  await page.route(urlPattern, async (route) => {
    await route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify(response),
    });
  });
}

/**
 * Helper function to create mock video
 */
export function createMockVideo(overrides: any = {}) {
  return {
    id: 'test-video-id',
    youtubeId: 'dQw4w9WgXcQ',
    title: 'Test Video',
    description: 'Test description',
    channelId: 'channel123',
    channelName: 'Test Channel',
    duration: 300,
    viewCount: 10000,
    likeCount: 500,
    publishedAt: '2024-01-01T00:00:00Z',
    thumbnailUrl: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg',
    tags: [],
    status: 'completed',
    hasTranscript: true,
    hasSummary: true,
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
    ...overrides,
  };
}

/**
 * Helper function to wait for toast notification
 */
export async function waitForToast(page: Page, message: string, timeout = 5000) {
  return page.waitForSelector(`text=${message}`, { timeout });
}

/**
 * Helper function to fill and submit form
 */
export async function fillAndSubmit(
  page: Page,
  inputSelector: string,
  value: string,
  submitSelector: string
) {
  await page.fill(inputSelector, value);
  await page.click(submitSelector);
}

