import { expect, test } from '@playwright/test';

const apiUrl = 'http://127.0.0.1:8080';

test('loads the app shell', async ({ page }) => {
  await page.goto('/');

  await expect(
    page.getByRole('heading', { name: 'Marketplace Assistant' })
  ).toBeVisible();
  await expect(page.getByRole('link', { name: 'Items' })).toBeVisible();
});

test('renders inventory items returned by the backend', async ({
  page,
  request
}) => {
  const title = `E2E denim jacket ${Date.now()}`;

  const response = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      description: 'Created by Playwright',
      category: 'Clothing',
      size: 'M',
      condition: 'Good',
      price_cents: 3200,
      currency: 'USD',
      notes: 'Visible in the inventory list'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/items');

  await expect(page.getByRole('heading', { name: 'Items' })).toBeVisible();
  await expect(page.getByText(title)).toBeVisible();
  await expect(
    page.locator('article').filter({ hasText: title }).getByText('$32.00')
  ).toBeVisible();
});

test('filters inventory by search text', async ({ page, request }) => {
  const title = `E2E leather boots ${Date.now()}`;

  const response = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      description: 'Created by Playwright',
      category: 'Shoes',
      size: '8',
      condition: 'Excellent',
      price_cents: 4500,
      currency: 'USD',
      notes: 'Searchable test item'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/items');
  await page.getByTestId('inventory-search').fill('leather boots');

  await expect(page.getByText(title)).toBeVisible();
  await expect(page.getByText('No inventory items')).toBeHidden();
});
