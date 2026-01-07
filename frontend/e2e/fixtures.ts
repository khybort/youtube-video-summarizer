import { test as base } from '@playwright/test';
import type { Page } from '@playwright/test';

// Extend base test with custom fixtures
export const test = base.extend<{
  mockVideo: any;
  mockTranscript: any;
  mockSummary: any;
}>({
  mockVideo: async ({ page }, use) => {
    const video = {
      id: 'test-video-id',
      youtubeId: 'dQw4w9WgXcQ',
      title: 'Test Video',
      description: 'Test description',
      channelName: 'Test Channel',
      viewCount: 10000,
      likeCount: 500,
      duration: 300,
      status: 'completed',
      hasTranscript: true,
      hasSummary: true,
    };
    await use(video);
  },

  mockTranscript: async ({ page }, use) => {
    const transcript = {
      id: 'transcript-id',
      videoId: 'test-video-id',
      language: 'en',
      source: 'whisper',
      content: 'This is a test transcript.',
      segments: [
        { start: 0, end: 5, text: 'This is a test transcript.' },
      ],
    };
    await use(transcript);
  },

  mockSummary: async ({ page }, use) => {
    const summary = {
      id: 'summary-id',
      videoId: 'test-video-id',
      modelUsed: 'gemini-1.5-flash',
      summaryType: 'detailed',
      content: 'This is a test summary.',
      keyPoints: ['Key point 1', 'Key point 2'],
    };
    await use(summary);
  },
});

export { expect } from '@playwright/test';

