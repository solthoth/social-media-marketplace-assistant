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

  it('lists item photos', () => {
    const { client, http } = setup();

    client.listItemPhotos('item-1').subscribe((response) => {
      expect(response.photos).toHaveLength(1);
      expect(response.photos[0].filename).toBe('front.png');
    });

    const request = http.expectOne('/api/items/item-1/photos');
    expect(request.request.method).toBe('GET');
    request.flush({ photos: [photoFixture()] });

    http.verify();
  });

  it('uploads an item photo as multipart form data', () => {
    const { client, http } = setup();
    const file = new File(['image-bytes'], 'front.png', { type: 'image/png' });

    client.uploadItemPhoto('item-1', file).subscribe((photo) => {
      expect(photo.id).toBe('photo-1');
    });

    const request = http.expectOne('/api/items/item-1/photos');
    expect(request.request.method).toBe('POST');
    expect(request.request.body instanceof FormData).toBe(true);
    expect(request.request.body.get('photo')).toBe(file);
    request.flush(photoFixture());

    http.verify();
  });

  it('reorders item photos', () => {
    const { client, http } = setup();

    client.reorderItemPhotos('item-1', ['photo-2', 'photo-1']).subscribe();

    const request = http.expectOne('/api/items/item-1/photos/order');
    expect(request.request.method).toBe('PATCH');
    expect(request.request.body).toEqual({
      photo_ids: ['photo-2', 'photo-1']
    });
    request.flush({
      photos: [photoFixture({ id: 'photo-2' }), photoFixture()]
    });

    http.verify();
  });

  it('sets an item photo as primary', () => {
    const { client, http } = setup();

    client.setPrimaryItemPhoto('item-1', 'photo-2').subscribe();

    const request = http.expectOne('/api/items/item-1/photos/photo-2/primary');
    expect(request.request.method).toBe('PATCH');
    request.flush({
      photos: [photoFixture({ id: 'photo-2', is_primary: true })]
    });

    http.verify();
  });

  it('deletes an item photo', () => {
    const { client, http } = setup();

    client.deleteItemPhoto('item-1', 'photo-1').subscribe();

    const request = http.expectOne('/api/items/item-1/photos/photo-1');
    expect(request.request.method).toBe('DELETE');
    request.flush(null);

    http.verify();
  });

  it('creates an item enrichment job', () => {
    const { client, http } = setup();

    client.createItemEnrichmentJob('item-1').subscribe((job) => {
      expect(job.id).toBe('job-1');
      expect(job.status).toBe('queued');
    });

    const request = http.expectOne('/api/items/item-1/enrichment-jobs');
    expect(request.request.method).toBe('POST');
    request.flush(enrichmentJobFixture());

    http.verify();
  });

  it('requests one item enrichment job', () => {
    const { client, http } = setup();

    client.getItemEnrichmentJob('item-1', 'job-1').subscribe((job) => {
      expect(job.suggestion.category).toBe('Clothing');
    });

    const request = http.expectOne('/api/items/item-1/enrichment-jobs/job-1');
    expect(request.request.method).toBe('GET');
    request.flush(enrichmentJobFixture({ status: 'completed' }));

    http.verify();
  });

  it('applies an item enrichment job', () => {
    const { client, http } = setup();

    client.applyItemEnrichmentJob('item-1', 'job-1').subscribe((response) => {
      expect(response.item.description).toBe('AI description');
      expect(response.applied_fields).toEqual(['description']);
    });

    const request = http.expectOne(
      '/api/items/item-1/enrichment-jobs/job-1/apply'
    );
    expect(request.request.method).toBe('POST');
    request.flush({
      item: itemFixture({ description: 'AI description' }),
      job: enrichmentJobFixture({ status: 'completed' }),
      applied_fields: ['description']
    });

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

function photoFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 'photo-1',
    item_id: 'item-1',
    filename: 'front.png',
    mime_type: 'image/png',
    sort_order: 0,
    is_primary: true,
    content_urls: {
      original: '/items/item-1/photos/photo-1/content?variant=original',
      medium: '/items/item-1/photos/photo-1/content?variant=medium',
      thumbnail: '/items/item-1/photos/photo-1/content?variant=thumbnail'
    },
    created_at: '2026-05-18T00:00:00Z',
    ...overrides
  };
}

function enrichmentJobFixture(overrides: Record<string, unknown> = {}) {
  return {
    id: 'job-1',
    item_id: 'item-1',
    status: 'queued',
    provider: 'fake',
    model: 'fake-vision',
    requested_at: '2026-05-18T00:00:00Z',
    started_at: null,
    completed_at: null,
    applied_at: null,
    error_message: '',
    input_snapshot: {
      item_id: 'item-1',
      title: 'Denim jacket',
      existing_description: '',
      existing_category: '',
      existing_size: '',
      existing_condition: '',
      existing_notes: '',
      photos: []
    },
    suggestion: {
      description: 'AI description',
      category: 'Clothing',
      size: 'M',
      condition: 'Good',
      notes: 'Review generated details.'
    },
    ...overrides
  };
}
