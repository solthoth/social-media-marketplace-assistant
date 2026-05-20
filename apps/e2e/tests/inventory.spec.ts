import { expect, test, type Locator, type Page } from '@playwright/test';
import { writeFileSync } from 'node:fs';

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
      original_purchase_price_cents: 1800,
      selling_price_cents: 3200,
      currency: 'USD',
      notes: 'Visible in the inventory list'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/items');

  await expect(page.getByRole('heading', { name: 'Items' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'New item' })).toHaveCount(1);
  await expect(page.getByText(title)).toBeVisible();
  await expect(
    page.locator('mat-card').filter({ hasText: title }).getByText('$18.00')
  ).toBeVisible();
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
      original_purchase_price_cents: 2800,
      selling_price_cents: 4500,
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
  await page.getByTestId('item-original-purchase-price').fill('9.25');
  await page.getByTestId('item-selling-price').fill('18.50');
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
      original_purchase_price_cents: 4100,
      selling_price_cents: 6400,
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
  await page.getByTestId('item-original-purchase-price').fill('45');
  await page.getByTestId('item-selling-price').fill('72');
  await page.getByTestId('save-draft').click();

  await expect(page).toHaveURL(/\/items$/);
  await expect(page.getByText(updatedTitle)).toBeVisible();
  await expect(
    page
      .locator('mat-card')
      .filter({ hasText: updatedTitle })
      .getByText('$45.00')
  ).toBeVisible();
  await expect(
    page
      .locator('mat-card')
      .filter({ hasText: updatedTitle })
      .getByText('$72.00')
  ).toBeVisible();
});

test('changes item status through the edit workflow controls', async ({
  page,
  request
}) => {
  const title = `E2E workflow handbag ${Date.now()}`;

  const response = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      category: 'Accessories',
      selling_price_cents: 3900,
      currency: 'USD'
    }
  });
  expect(response.ok()).toBeTruthy();

  await page.goto('/items');
  const card = page.locator('mat-card').filter({ hasText: title });
  const statusBadge = card.locator('.status-badge');
  await expect(statusBadge).toHaveText('Draft');
  await expect(card.getByTestId(/item-status-/)).toHaveCount(0);

  await changeStatusFromEditPage(page, card, 'ready_to_list');
  await expect(statusBadge).toHaveText('Ready to list');

  await changeStatusFromEditPage(page, card, 'listed');
  await expect(statusBadge).toHaveText('Listed');

  await changeStatusFromEditPage(page, card, 'sold');
  await expect(statusBadge).toHaveText('Sold');

  await changeStatusFromEditPage(page, card, 'listed');
  await expect(statusBadge).toHaveText('Listed');

  await changeStatusFromEditPage(page, card, 'archived');
  await expect(statusBadge).toHaveText('Archived');

  await changeStatusFromEditPage(page, card, 'draft');
  await expect(statusBadge).toHaveText('Draft');
});

test('uploads, previews, reorders, marks primary, and removes item photos', async ({
  page,
  request
}, testInfo) => {
  const title = `E2E photo blouse ${Date.now()}`;
  const create = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      category: 'Clothing',
      selling_price_cents: 2800,
      currency: 'USD'
    }
  });
  expect(create.ok()).toBeTruthy();
  const item = await create.json();

  const frontPhoto = testInfo.outputPath('front.png');
  const backPhoto = testInfo.outputPath('back.png');
  writeFileSync(frontPhoto, Buffer.from(pngBytes));
  writeFileSync(backPhoto, Buffer.from(pngBytes));

  await page.goto(`/items/${item.id}/edit`);
  await expect(page.getByRole('heading', { name: 'Edit item' })).toBeVisible();
  await expect(page.getByText('No photos yet')).toBeVisible();

  await page.getByTestId('item-photo-input').setInputFiles(frontPhoto);
  const frontCard = page
    .locator('.photo-card')
    .filter({ hasText: 'front.png' });
  await expect(frontCard).toBeVisible();
  await expect(frontCard.getByRole('img', { name: 'front.png' })).toBeVisible();
  await expect(frontCard.locator('.status-badge')).toHaveText('Primary');

  await page.getByTestId('item-photo-input').setInputFiles(backPhoto);
  const backCard = page.locator('.photo-card').filter({ hasText: 'back.png' });
  await expect(backCard).toBeVisible();

  await backCard.getByRole('button', { name: 'Primary' }).click();
  await expect(backCard.locator('.status-badge')).toHaveText('Primary');
  await expect(frontCard.locator('.status-badge')).toHaveCount(0);

  await backCard.getByRole('button', { name: 'Up' }).click();
  await expect(
    page.locator('[data-testid="photo-filename"]').first()
  ).toHaveText('back.png');

  await frontCard.getByRole('button', { name: 'Remove' }).click();
  await expect(frontCard).toHaveCount(0);
  await expect(backCard).toBeVisible();
});

test('generates item details from title and photos with AI assist', async ({
  page,
  request
}, testInfo) => {
  const title = `E2E AI denim jacket ${Date.now()}`;
  const create = await request.post(`${apiUrl}/items`, {
    data: {
      title,
      currency: 'USD'
    }
  });
  expect(create.ok()).toBeTruthy();
  const item = await create.json();

  const photo = testInfo.outputPath('ai-front.png');
  writeFileSync(photo, Buffer.from(pngBytes));

  await page.goto(`/items/${item.id}/edit`);
  await page.getByTestId('item-photo-input').setInputFiles(photo);
  await expect(
    page.locator('.photo-card').filter({ hasText: 'ai-front.png' })
  ).toBeVisible();

  await page.getByTestId('generate-details').click();

  await expect(page.getByTestId('item-description')).toHaveValue(
    /AI-generated draft details/
  );
  await expect(page.getByTestId('item-category')).toHaveValue('Uncategorized');
  await expect(page.getByText(/Filled description, category/)).toBeVisible();
});

async function changeStatusFromEditPage(
  page: Page,
  card: Locator,
  status: string
) {
  await card.getByRole('link', { name: 'Edit' }).click();
  await page.getByTestId('item-status').selectOption(status);
  await page.getByTestId('save-draft').click();
  await expect(page).toHaveURL(/\/items$/);
}

const pngBytes = [
  0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d, 0x49,
  0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x08, 0x02,
  0x00, 0x00, 0x00, 0x90, 0x77, 0x53, 0xde, 0x00, 0x00, 0x00, 0x0c, 0x49, 0x44,
  0x41, 0x54, 0x08, 0xd7, 0x63, 0xf8, 0xcf, 0xc0, 0x00, 0x00, 0x03, 0x01, 0x01,
  0x00, 0x18, 0xdd, 0x8d, 0xb0, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44,
  0xae, 0x42, 0x60, 0x82
];
