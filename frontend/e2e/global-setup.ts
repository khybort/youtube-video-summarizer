import { chromium, FullConfig } from '@playwright/test';

async function globalSetup(config: FullConfig) {
  // Optional: Setup tasks before all tests
  // For example, seed test data, start services, etc.
  
  try {
    const browser = await chromium.launch();
    const page = await browser.newPage();
    
    // Wait for the app to be ready
    try {
      await page.goto(config.use?.baseURL || 'http://localhost:3000', { 
        waitUntil: 'domcontentloaded',
        timeout: 10000 
      });
      await page.waitForLoadState('networkidle', { timeout: 5000 });
    } catch (error) {
      console.warn('App might not be ready, continuing anyway:', error);
    }
    
    await browser.close();
  } catch (error) {
    console.warn('Global setup failed, continuing anyway:', error);
  }
}

export default globalSetup;

