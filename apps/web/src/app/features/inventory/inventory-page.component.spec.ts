import { TestBed } from '@angular/core/testing';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { describe, expect, it } from 'vitest';
import { InventoryPageComponent } from './inventory-page.component';

describe('InventoryPageComponent', () => {
  function setup() {
    TestBed.configureTestingModule({
      imports: [InventoryPageComponent],
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        provideRouter([])
      ]
    });

    const fixture = TestBed.createComponent(InventoryPageComponent);
    const http = TestBed.inject(HttpTestingController);
    fixture.detectChanges();

    return { fixture, http };
  }

  it('loads and renders inventory items', () => {
    const { fixture, http } = setup();

    flushItems(http, [
      itemFixture({
        id: 'item-1',
        title: 'Denim jacket',
        category: 'Clothing',
        status: 'draft',
        original_purchase_price_cents: 1800,
        selling_price_cents: 3200
      }),
      itemFixture({
        id: 'item-2',
        title: 'Leather boots',
        category: 'Shoes',
        status: 'ready_to_list',
        original_purchase_price_cents: 2400,
        selling_price_cents: 4200
      })
    ]);
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Denim jacket');
    expect(text).toContain('Leather boots');
    expect(text).toContain('Original price');
    expect(text).toContain('$18.00');
    expect(text).toContain('Selling price');
    expect(text).toContain('$32.00');
    expect(text).toContain('2 items');
    expect(fixture.nativeElement.querySelectorAll('mat-card')).toHaveLength(2);
    expect(
      fixture.nativeElement.querySelector('a[href="/items/new"]')
    ).toBeFalsy();
    expect(
      fixture.nativeElement.querySelector('a[href="/items/item-1/edit"]')
    ).toBeTruthy();
    expect(
      fixture.nativeElement.querySelector('[data-testid="item-status-item-1"]')
    ).toBeFalsy();
  });

  it('filters loaded items by search text', () => {
    const { fixture, http } = setup();

    flushItems(http, [
      itemFixture({ title: 'Denim jacket', category: 'Clothing' }),
      itemFixture({ title: 'Leather boots', category: 'Shoes' })
    ]);
    fixture.detectChanges();

    const input: HTMLInputElement = fixture.nativeElement.querySelector(
      '[data-testid="inventory-search"]'
    );
    input.value = 'boots';
    input.dispatchEvent(new Event('input'));
    fixture.detectChanges();

    const text = fixture.nativeElement.textContent;
    expect(text).toContain('Leather boots');
    expect(text).not.toContain('Denim jacket');
  });

  it('reloads items when status filter changes', () => {
    const { fixture, http } = setup();

    flushItems(http, [itemFixture({ title: 'Denim jacket' })]);
    fixture.detectChanges();

    const select: HTMLSelectElement = fixture.nativeElement.querySelector(
      '[data-testid="status-filter"]'
    );
    select.value = 'ready_to_list';
    select.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === '/api/items' &&
        candidate.params.get('status') === 'ready_to_list'
    );
    expect(request.request.method).toBe('GET');
    request.flush({
      items: [itemFixture({ title: 'Leather boots', status: 'ready_to_list' })]
    });
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Leather boots');
    http.verify();
  });

  it('does not expose item status editing from the inventory list', () => {
    const { fixture, http } = setup();

    flushItems(http, [itemFixture({ id: 'item-1', status: 'draft' })]);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Draft');
    expect(
      fixture.nativeElement.querySelector('[data-testid="item-status-item-1"]')
    ).toBeFalsy();
    http.verify();
  });

  it('shows an empty state when no items match', () => {
    const { fixture, http } = setup();

    flushItems(http, []);
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('No inventory items');
  });
});

function flushItems(http: HttpTestingController, items: unknown[]) {
  const request = http.expectOne('/api/items');
  expect(request.request.method).toBe('GET');
  request.flush({ items });
  http.verify();
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
