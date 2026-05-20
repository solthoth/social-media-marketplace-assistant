import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { InventoryItem } from '../../core/api-client.service';
import { ItemAiAssistComponent } from './item-ai-assist.component';

describe('ItemAiAssistComponent', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('disables generation until the item has a title and photos', () => {
    const { fixture, http } = setup({ title: '', photoCount: 0 });

    const button = generateButton(fixture);

    expect(button.disabled).toBe(true);
    expect(fixture.nativeElement.textContent).toContain(
      'Add a title and at least one photo'
    );
    http.verify();
  });

  it('creates, polls, applies, and emits completed enrichment results', async () => {
    vi.useFakeTimers();
    const { fixture, http, appliedItems } = setup({
      title: 'Denim jacket',
      photoCount: 1
    });

    generateButton(fixture).click();
    fixture.detectChanges();

    const create = http.expectOne('/api/items/item-1/enrichment-jobs');
    expect(create.request.method).toBe('POST');
    create.flush(enrichmentJobFixture());

    await vi.advanceTimersByTimeAsync(1000);
    const poll = http.expectOne('/api/items/item-1/enrichment-jobs/job-1');
    expect(poll.request.method).toBe('GET');
    poll.flush(enrichmentJobFixture({ status: 'completed' }));

    const apply = http.expectOne(
      '/api/items/item-1/enrichment-jobs/job-1/apply'
    );
    expect(apply.request.method).toBe('POST');
    apply.flush({
      item: itemFixture({
        description: 'AI-generated draft details for Denim jacket.',
        category: 'Uncategorized'
      }),
      job: enrichmentJobFixture({ status: 'completed' }),
      applied_fields: ['description', 'category']
    });
    fixture.detectChanges();

    expect(appliedItems[0].description).toContain('AI-generated');
    expect(fixture.nativeElement.textContent).toContain(
      'Filled description, category'
    );
    http.verify();
  });
});

function setup({ title, photoCount }: { title: string; photoCount: number }) {
  TestBed.configureTestingModule({
    imports: [ItemAiAssistComponent],
    providers: [provideHttpClient(), provideHttpClientTesting()]
  });

  const fixture = TestBed.createComponent(ItemAiAssistComponent);
  fixture.componentRef.setInput('itemId', 'item-1');
  fixture.componentRef.setInput('title', title);
  fixture.componentRef.setInput('photoCount', photoCount);
  const appliedItems: InventoryItem[] = [];
  fixture.componentInstance.itemApplied.subscribe((item) =>
    appliedItems.push(item)
  );
  const http = TestBed.inject(HttpTestingController);
  fixture.detectChanges();

  return { fixture, http, appliedItems };
}

function generateButton(fixture: ComponentFixture<ItemAiAssistComponent>) {
  return fixture.nativeElement.querySelector(
    '[data-testid="generate-details"]'
  ) as HTMLButtonElement;
}

function itemFixture(overrides: Partial<InventoryItem> = {}): InventoryItem {
  return {
    id: 'item-1',
    title: 'Denim jacket',
    description: '',
    category: '',
    size: '',
    condition: '',
    original_purchase_price_cents: 0,
    selling_price_cents: 0,
    currency: 'USD',
    status: 'draft',
    notes: '',
    created_at: '2026-05-18T00:00:00Z',
    updated_at: '2026-05-18T00:00:00Z',
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
    suggestion: {
      description: 'AI-generated draft details for Denim jacket.',
      category: 'Uncategorized',
      size: '',
      condition: '',
      notes: ''
    },
    ...overrides
  };
}
