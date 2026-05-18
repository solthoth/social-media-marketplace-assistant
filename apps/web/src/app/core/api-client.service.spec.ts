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
          price_cents: 3200,
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
});
