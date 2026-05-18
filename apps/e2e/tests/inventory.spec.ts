import { expect, test } from '@playwright/test';

const apiUrl = 'http://127.0.0.1:8080';

test('loads the app shell', async ({ page }) => {
  await page.goto('/');

  await expect(
    page.getByRole('heading', { name: 'Marketplace Assistant' })
  ).toBeVisible();
  await expect(page.getByRole('link', { name: 'Items' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'New item' })).toHaveAttribute(
    'href',
    '/items/new'
  );
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
    page.locator('mat-card').filter({ hasText: title }).getByText('$32.00')
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

test('creates a draft item from the item form', async ({ page }) => {
  const title = `E2E silk scarf ${Date.now()}`;

  await page.goto('/items/new');
  await page.getByTestId('item-title').fill(title);
  await page.getByTestId('item-description').fill('Created through the UI');
  await page.getByTestId('item-category').fill('Accessories');
  await page.getByTestId('item-size').fill('One size');
  await page.getByTestId('item-condition').fill('Excellent');
  await page.getByTestId('item-price').fill('18.50');
  await page.getByTestId('item-notes').fill('Fold neatly before listing');
  await page.getByTestId('save-draft').click();

  await expect(page).toHaveURL(/\/items$/);
  await expect(page.getByText(title)).toBeVisible();
  await expect(
    page.locator('mat-card').filter({ hasText: title }).getByText('$18.50')
  ).toBeVisible();
});

test('opens the item form from primary navigation', async ({ page }) => {
  await page.goto('/');
  await page.getByRole('link', { name: 'New item' }).click();

  await expect(page).toHaveURL(/\/items\/new$/);
  await expect(page.getByRole('heading', { name: 'New item' })).toBeVisible();
});

test('requires a title when saving the item form', async ({ page }) => {
  await page.goto('/items/new');

  await expect(page.getByRole('heading', { name: 'New item' })).toBeVisible();
  await page.getByTestId('save-draft').click();

  await expect(page.getByText('Title is required')).toBeVisible();
  await expect(page).toHaveURL(/\/items\/new$/);
});

test('edits an existing item from the inventory list', async ({
  page,
  request
}) => {
  const title = `E2E wool coat ${Date.now()}`;
  const updatedTitle = `${title} updated`;

  const response = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      description: 'Created before editing',
      category: 'Clothing',
      size: 'L',
      condition: 'Good',
      price_cents: 6400,
      currency: 'USD',
      notes: 'Needs lint roller'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/items');
  await page
    .locator('mat-card')
    .filter({ hasText: title })
    .getByRole('link', { name: 'Edit' })
    .click();

  await expect(page.getByRole('heading', { name: 'Edit item' })).toBeVisible();
  await expect(page.getByTestId('item-title')).toHaveValue(title);

  await page.getByTestId('item-title').fill(updatedTitle);
  await page.getByTestId('item-price').fill('72');
  await page.getByTestId('save-draft').click();

  await expect(page).toHaveURL(/\/items$/);
  await expect(page.getByText(updatedTitle)).toBeVisible();
  await expect(
    page
      .locator('mat-card')
      .filter({ hasText: updatedTitle })
      .getByText('$72.00')
  ).toBeVisible();
});
