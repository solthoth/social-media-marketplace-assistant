import { TestBed } from '@angular/core/testing';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { describe, expect, it } from 'vitest';
import { ApiClientService } from './api-client.service';

describe('ApiClientService', () => {
  function setup() {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        ApiClientService
      ]
    });

    return {
      client: TestBed.inject(ApiClientService),
      http: TestBed.inject(HttpTestingController)
    };
  }

  it('requests API health', () => {
    const { client, http } = setup();

    client.health().subscribe((response) => {
      expect(response.status).toBe('ok');
      expect(response.service).toBe('social-media-marketplace-assistant-api');
    });

    const request = http.expectOne('/api/healthz');
    expect(request.request.method).toBe('GET');
    request.flush({
      status: 'ok',
      service: 'social-media-marketplace-assistant-api',
      time: '2026-05-18T00:00:00Z'
    });

    http.verify();
  });

  it('requests inventory items without status filter', () => {
    const { client, http } = setup();

    client.listItems().subscribe((response) => {
      expect(response.items).toHaveLength(1);
      expect(response.items[0].title).toBe('Denim jacket');
    });

    const request = http.expectOne('/api/items');
    expect(request.request.method).toBe('GET');
    request.flush({
      items: [
        {
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
          updated_at: '2026-05-18T00:00:00Z'
        }
      ]
    });

    http.verify();
  });

  it('requests inventory items with status filter', () => {
    const { client, http } = setup();

    client.listItems('ready_to_list').subscribe();

    const request = http.expectOne(
      (candidate) =>
        candidate.url === '/api/items' &&
        candidate.params.get('status') === 'ready_to_list'
    );
    expect(request.request.method).toBe('GET');
    request.flush({ items: [] });

    http.verify();
  });

  it('creates an inventory item draft', () => {
    const { client, http } = setup();

    client
      .createItem({
        title: 'Denim jacket',
        description: 'Medium wash',
        category: 'Clothing',
        size: 'M',
        condition: 'Good',
        original_purchase_price_cents: 1800,
        selling_price_cents: 3200,
        currency: 'USD',
        notes: 'Steam before photos'
      })
      .subscribe((item) => {
        expect(item.id).toBe('item-1');
        expect(item.status).toBe('draft');
      });

    const request = http.expectOne('/api/items');
    expect(request.request.method).toBe('POST');
    expect(request.request.body).toEqual({
      title: 'Denim jacket',
      description: 'Medium wash',
      category: 'Clothing',
      size: 'M',
      condition: 'Good',
      original_purchase_price_cents: 1800,
      selling_price_cents: 3200,
      currency: 'USD',
      notes: 'Steam before photos'
    });
    request.flush(itemFixture());

    http.verify();
  });

  it('requests one inventory item', () => {
    const { client, http } = setup();

    client.getItem('item-1').subscribe((item) => {
      expect(item.title).toBe('Denim jacket');
    });

    const request = http.expectOne('/api/items/item-1');
    expect(request.request.method).toBe('GET');
    request.flush(itemFixture());

    http.verify();
  });

  it('updates an inventory item', () => {
    const { client, http } = setup();

    client
      .updateItem('item-1', {
        title: 'Updated denim jacket',
        original_purchase_price_cents: 1800,
        selling_price_cents: 3600,
        currency: 'USD'
      })
      .subscribe((item) => {
        expect(item.title).toBe('Updated denim jacket');
      });

    const request = http.expectOne('/api/items/item-1');
    expect(request.request.method).toBe('PATCH');
    expect(request.request.body).toEqual({
      title: 'Updated denim jacket',
      original_purchase_price_cents: 1800,
      selling_price_cents: 3600,
      currency: 'USD'
    });
    request.flush(itemFixture({ title: 'Updated denim jacket' }));

    http.verify();
  });
});

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
