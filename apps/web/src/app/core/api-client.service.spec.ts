import { TestBed } from '@angular/core/testing';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { provideHttpClient } from '@angular/common/http';
import { describe, expect, it } from 'vitest';
import { ApiClientService } from './api-client.service';

describe('ApiClientService', () => {
  it('requests API health', () => {
    TestBed.configureTestingModule({
      providers: [
        provideHttpClient(),
        provideHttpClientTesting(),
        ApiClientService
      ]
    });

    const client = TestBed.inject(ApiClientService);
    const http = TestBed.inject(HttpTestingController);

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
});
