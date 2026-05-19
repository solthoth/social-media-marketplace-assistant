import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import {
  ActivatedRoute,
  convertToParamMap,
  provideRouter,
  Router
} from '@angular/router';
import { describe, expect, it } from 'vitest';
import { ItemFormPageComponent } from './item-form-page.component';

describe('ItemFormPageComponent', () => {
  it('validates title before saving a draft', () => {
    const { fixture, http } = setup('/items/new');

    clickSave(fixture);

    expect(fixture.nativeElement.textContent).toContain('Title is required');
    expect(
      fixture.nativeElement.querySelectorAll('mat-form-field').length
    ).toBe(9);
    const currency: HTMLSelectElement = fixture.nativeElement.querySelector(
      '[data-testid="item-currency"]'
    );
    expect(Array.from(currency.options).map((option) => option.value)).toEqual([
      'USD'
    ]);
    http.expectNone('/api/items');
    http.verify();
  });

  it('creates a draft item from form values', async () => {
    const { fixture, http, router } = setup('/items/new');

    fillInput(fixture, 'item-title', 'Denim jacket');
    fillInput(fixture, 'item-description', 'Medium wash');
    fillInput(fixture, 'item-category', 'Clothing');
    fillInput(fixture, 'item-size', 'M');
    fillInput(fixture, 'item-condition', 'Good');
    fillInput(fixture, 'item-original-purchase-price', '18');
    fillInput(fixture, 'item-selling-price', '32.50');
    fillInput(fixture, 'item-notes', 'Steam before photos');
    clickSave(fixture);

    const request = http.expectOne('/api/items');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({
      title: 'Denim jacket',
      description: 'Medium wash',
      category: 'Clothing',
      size: 'M',
      condition: 'Good',
      original_purchase_price_cents: 1800,
      selling_price_cents: 3250,
      currency: 'USD',
      notes: 'Steam before photos'
    });
    request.flush(itemFixture());
    fixture.detectChanges();
    await fixture.whenStable();

    expect(router.url).toBe('/items');
    http.verify();
  });

  it('loads an existing item and patches edits', async () => {
    const { fixture, http, router } = setup('/items/item-1/edit');

    const getRequest = http.expectOne('/api/items/item-1');
    expect(getRequest.request.method).toBe('GET');
    getRequest.flush(itemFixture());
    fixture.detectChanges();

    const title: HTMLInputElement = fixture.nativeElement.querySelector(
      '[data-testid="item-title"]'
    );
    expect(title.value).toBe('Denim jacket');

    fillInput(fixture, 'item-title', 'Updated denim jacket');
    fillInput(fixture, 'item-original-purchase-price', '20');
    fillInput(fixture, 'item-selling-price', '36');
    clickSave(fixture);

    const patchRequest = http.expectOne('/api/items/item-1');
    expect(patchRequest.request.method).toBe('PATCH');
    expect(patchRequest.request.body.title).toBe('Updated denim jacket');
    expect(patchRequest.request.body.original_purchase_price_cents).toBe(2000);
    expect(patchRequest.request.body.selling_price_cents).toBe(3600);
    expect(patchRequest.request.body.status).toBe('draft');
    patchRequest.flush(itemFixture({ title: 'Updated denim jacket' }));
    fixture.detectChanges();
    await fixture.whenStable();

    expect(router.url).toBe('/items');
    http.verify();
  });

  it('allows valid status changes while editing an item', async () => {
    const { fixture, http, router } = setup('/items/item-1/edit');

    const getRequest = http.expectOne('/api/items/item-1');
    getRequest.flush(itemFixture({ status: 'ready_to_list' }));
    fixture.detectChanges();

    const status: HTMLSelectElement = fixture.nativeElement.querySelector(
      '[data-testid="item-status"]'
    );
    expect(Array.from(status.options).map((option) => option.value)).toEqual([
      'draft',
      'ready_to_list',
      'listed',
      'archived'
    ]);

    status.value = 'listed';
    status.dispatchEvent(new Event('change'));
    fixture.detectChanges();
    clickSave(fixture);

    const patchRequest = http.expectOne('/api/items/item-1');
    expect(patchRequest.request.method).toBe('PATCH');
    expect(patchRequest.request.body.status).toBe('listed');
    patchRequest.flush(itemFixture({ status: 'listed' }));
    fixture.detectChanges();
    await fixture.whenStable();

    expect(router.url).toBe('/items');
    http.verify();
  });
});

function setup(initialUrl: string) {
  const match = initialUrl.match(/^\/items\/([^/]+)\/edit$/);

  TestBed.configureTestingModule({
    imports: [ItemFormPageComponent],
    providers: [
      provideHttpClient(),
      provideHttpClientTesting(),
      provideRouter([{ path: 'items', component: ItemFormPageComponent }]),
      {
        provide: ActivatedRoute,
        useValue: {
          snapshot: {
            paramMap: convertToParamMap(match ? { id: match[1] } : {})
          }
        }
      }
    ]
  });

  const router = TestBed.inject(Router);
  const fixture = TestBed.createComponent(ItemFormPageComponent);
  const http = TestBed.inject(HttpTestingController);
  fixture.detectChanges();

  return { fixture, http, router };
}

function fillInput(
  fixture: ComponentFixture<ItemFormPageComponent>,
  testId: string,
  value: string
) {
  const input: HTMLInputElement | HTMLTextAreaElement =
    fixture.nativeElement.querySelector(`[data-testid="${testId}"]`);
  input.value = value;
  input.dispatchEvent(new Event('input'));
  fixture.detectChanges();
}

function clickSave(fixture: ComponentFixture<ItemFormPageComponent>) {
  const button: HTMLButtonElement = fixture.nativeElement.querySelector(
    '[data-testid="save-draft"]'
  );
  button.click();
  fixture.detectChanges();
}

function itemFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 'item-1',
    title: 'Denim jacket',
    description: 'Medium wash',
    category: 'Clothing',
    size: 'M',
    condition: 'Good',
    original_purchase_price_cents: 1800,
    selling_price_cents: 3200,
    currency: 'USD',
    status: 'draft',
    notes: 'Steam before photos',
    created_at: '2026-05-18T00:00:00Z',
    updated_at: '2026-05-18T00:00:00Z',
    ...overrides
  };
}
