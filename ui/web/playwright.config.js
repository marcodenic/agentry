import { defineConfig } from '@playwright/test';
export default defineConfig({
  testDir: './tests',
  webServer: {
    command: 'go run ../../cmd/agentry serve --config ../../examples/.agentry.yaml --metrics',
    url: 'http://localhost:8080',
    reuseExistingServer: true,
    timeout: 60000
  }
});
