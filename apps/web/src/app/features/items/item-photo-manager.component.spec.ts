import { provideHttpClient } from '@angular/common/http';
import {
  HttpTestingController,
  provideHttpClientTesting
} from '@angular/common/http/testing';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { describe, expect, it } from 'vitest';
import { ItemPhotoManagerComponent } from './item-photo-manager.component';

describe('ItemPhotoManagerComponent', () => {
  it('loads and previews existing photos', () => {
    const { fixture, http } = setup();

    http.expectOne('/api/items/item-1/photos').flush({
      photos: [photoFixture()]
    });
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('front.png');
    expect(fixture.nativeElement.textContent).toContain('Primary');
    const image: HTMLImageElement = fixture.nativeElement.querySelector(
      '[data-testid="photo-preview-photo-1"]'
    );
    expect(image.src).toContain(
      '/api/items/item-1/photos/photo-1/content?variant=thumbnail'
    );
    http.verify();
  });

  it('uploads selected photo files and refreshes the list', () => {
    const { fixture, http } = setup();
    http.expectOne('/api/items/item-1/photos').flush({ photos: [] });
    fixture.detectChanges();

    const file = new File(['image-bytes'], 'front.png', {
      type: 'image/png'
    });
    const input: HTMLInputElement = fixture.nativeElement.querySelector(
      '[data-testid="item-photo-input"]'
    );
    Object.defineProperty(input, 'files', {
      value: [file],
      configurable: true
    });
    input.dispatchEvent(new Event('change'));
    fixture.detectChanges();

    const upload = http.expectOne('/api/items/item-1/photos');
    expect(upload.request.method).toBe('POST');
    expect(upload.request.body.get('photo')).toBe(file);
    upload.flush(photoFixture());
    http.expectOne('/api/items/item-1/photos').flush({
      photos: [photoFixture()]
    });
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('front.png');
    http.verify();
  });

  it('removes photos after confirmation from the API', () => {
    const { fixture, http } = setup();
    http.expectOne('/api/items/item-1/photos').flush({
      photos: [photoFixture()]
    });
    fixture.detectChanges();

    click(fixture, 'remove-photo-photo-1');

    const remove = http.expectOne('/api/items/item-1/photos/photo-1');
    expect(remove.request.method).toBe('DELETE');
    remove.flush(null);
    http.expectOne('/api/items/item-1/photos').flush({ photos: [] });
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('No photos yet');
    http.verify();
  });

  it('reorders photos with move controls', () => {
    const { fixture, http } = setup();
    http.expectOne('/api/items/item-1/photos').flush({
      photos: [
        photoFixture(),
        photoFixture({ id: 'photo-2', filename: 'back.png', sort_order: 1 })
      ]
    });
    fixture.detectChanges();

    click(fixture, 'move-photo-up-photo-2');

    const reorder = http.expectOne('/api/items/item-1/photos/order');
    expect(reorder.request.method).toBe('PATCH');
    expect(reorder.request.body).toEqual({
      photo_ids: ['photo-2', 'photo-1']
    });
    reorder.flush({
      photos: [
        photoFixture({ id: 'photo-2', filename: 'back.png', sort_order: 0 }),
        photoFixture({ sort_order: 1 })
      ]
    });
    fixture.detectChanges();

    const filenames = Array.from<HTMLElement>(
      fixture.nativeElement.querySelectorAll('[data-testid="photo-filename"]')
    ).map((element) => element.textContent.trim());
    expect(filenames).toEqual(['back.png', 'front.png']);
    http.verify();
  });

  it('marks a photo as primary through the API', () => {
    const { fixture, http } = setup();
    http.expectOne('/api/items/item-1/photos').flush({
      photos: [
        photoFixture(),
        photoFixture({
          id: 'photo-2',
          filename: 'back.png',
          sort_order: 1,
          is_primary: false
        })
      ]
    });
    fixture.detectChanges();

    click(fixture, 'set-primary-photo-photo-2');

    const primary = http.expectOne('/api/items/item-1/photos/photo-2/primary');
    expect(primary.request.method).toBe('PATCH');
    primary.flush({
      photos: [
        photoFixture({ is_primary: false }),
        photoFixture({
          id: 'photo-2',
          filename: 'back.png',
          sort_order: 1,
          is_primary: true
        })
      ]
    });
    fixture.detectChanges();

    expect(fixture.nativeElement.textContent).toContain('Primary');
    http.verify();
  });
});

function setup() {
  TestBed.configureTestingModule({
    imports: [ItemPhotoManagerComponent],
    providers: [provideHttpClient(), provideHttpClientTesting()]
  });

  const fixture = TestBed.createComponent(ItemPhotoManagerComponent);
  fixture.componentRef.setInput('itemId', 'item-1');
  const http = TestBed.inject(HttpTestingController);
  fixture.detectChanges();

  return { fixture, http };
}

function click(
  fixture: ComponentFixture<ItemPhotoManagerComponent>,
  testId: string
) {
  const button: HTMLButtonElement = fixture.nativeElement.querySelector(
    `[data-testid="${testId}"]`
  );
  button.click();
  fixture.detectChanges();
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
