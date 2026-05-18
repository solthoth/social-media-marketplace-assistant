import { defineConfig, devices } from '@playwright/test';

const apiUrl = 'http://127.0.0.1:8080';
const webUrl = 'http://127.0.0.1:4200';
const databasePath = `/tmp/social-media-marketplace-assistant-e2e-${process.pid}.db`;

export default defineConfig({
  testDir: './tests',
  timeout: 30_000,
  expect: {
    timeout: 5_000
  },
  fullyParallel: false,
  workers: 1,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: webUrl,
    trace: 'on-first-retry'
  },
  webServer: [
    {
      command: `DATABASE_PATH=${databasePath} PORT=8080 go run ./services/api/cmd/api`,
      cwd: '../..',
      url: `${apiUrl}/healthz`,
      reuseExistingServer: false,
      timeout: 60_000
    },
    {
      command: 'npm --workspace apps/web start',
      cwd: '../..',
      url: webUrl,
      reuseExistingServer: false,
      timeout: 60_000
    }
  ],
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] }
    }
  ]
});
