import { NgFor, NgIf } from '@angular/common';
import { Component, inject, input, OnInit, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { ApiClientService, ItemPhoto } from '../../core/api-client.service';

@Component({
  selector: 'smm-item-photo-manager',
  standalone: true,
  imports: [MatButtonModule, MatCardModule, NgFor, NgIf],
  template: `
    <section class="photo-manager" aria-labelledby="photo-manager-title">
      <div class="photo-manager-header">
        <div>
          <h2 id="photo-manager-title">Photos</h2>
          <p class="photo-manager-summary">
            Add camera photos or image files, then arrange them for listing.
          </p>
        </div>
        <button
          matButton="filled"
          class="primary-action button-action"
          type="button"
          data-testid="choose-photo"
          [disabled]="isUploading()"
          (click)="fileInput.click()"
        >
          Add photos
        </button>
        <input
          #fileInput
          class="photo-input"
          data-testid="item-photo-input"
          type="file"
          accept="image/jpeg,image/png,image/webp"
          capture="environment"
          multiple
          (change)="uploadSelectedPhotos($event)"
        />
      </div>

      <p *ngIf="loadError()" class="notice error" role="alert">
        Photos could not be loaded.
      </p>
      <p *ngIf="uploadError()" class="notice error" role="alert">
        Photo could not be uploaded.
      </p>

      <p *ngIf="isLoading()" class="notice">Loading photos...</p>
      <p *ngIf="!isLoading() && photos().length === 0" class="empty-state">
        No photos yet.
      </p>

      <div *ngIf="!isLoading() && photos().length > 0" class="photo-grid">
        <mat-card
          *ngFor="let photo of photos(); let index = index"
          appearance="outlined"
          class="photo-card"
        >
          <img
            [attr.data-testid]="'photo-preview-' + photo.id"
            [src]="photoPreviewURL(photo)"
            [alt]="photo.filename"
            loading="lazy"
          />
          <mat-card-content>
            <div class="photo-meta">
              <strong data-testid="photo-filename">{{ photo.filename }}</strong>
              <span *ngIf="photo.is_primary" class="status-badge">Primary</span>
            </div>
          </mat-card-content>
          <mat-card-actions align="end">
            <button
              matButton="outlined"
              class="secondary-action button-action compact-action"
              type="button"
              [attr.data-testid]="'move-photo-up-' + photo.id"
              [disabled]="index === 0 || isBusy()"
              (click)="movePhoto(index, -1)"
            >
              Up
            </button>
            <button
              matButton="outlined"
              class="secondary-action button-action compact-action"
              type="button"
              [attr.data-testid]="'move-photo-down-' + photo.id"
              [disabled]="index === photos().length - 1 || isBusy()"
              (click)="movePhoto(index, 1)"
            >
              Down
            </button>
            <button
              matButton="outlined"
              class="secondary-action button-action compact-action"
              type="button"
              [attr.data-testid]="'set-primary-photo-' + photo.id"
              [disabled]="photo.is_primary || isBusy()"
              (click)="setPrimary(photo)"
            >
              Primary
            </button>
            <button
              matButton="outlined"
              class="secondary-action button-action compact-action danger-action"
              type="button"
              [attr.data-testid]="'remove-photo-' + photo.id"
              [disabled]="isBusy()"
              (click)="removePhoto(photo)"
            >
              Remove
            </button>
          </mat-card-actions>
        </mat-card>
      </div>
    </section>
  `
})
export class ItemPhotoManagerComponent implements OnInit {
  readonly itemId = input.required<string>();

  private readonly api = inject(ApiClientService);

  protected readonly photos = signal<ItemPhoto[]>([]);
  protected readonly isLoading = signal(true);
  protected readonly isUploading = signal(false);
  protected readonly isUpdating = signal(false);
  protected readonly loadError = signal(false);
  protected readonly uploadError = signal(false);

  ngOnInit(): void {
    this.loadPhotos();
  }

  protected isBusy(): boolean {
    return this.isUploading() || this.isUpdating();
  }

  protected uploadSelectedPhotos(event: Event): void {
    const inputElement = event.target as HTMLInputElement;
    const files = Array.from(inputElement.files ?? []);
    inputElement.value = '';
    if (files.length === 0) {
      return;
    }

    this.isUploading.set(true);
    this.uploadError.set(false);
    this.uploadNext(files, 0);
  }

  protected movePhoto(index: number, direction: -1 | 1): void {
    const nextIndex = index + direction;
    const current = [...this.photos()];
    if (nextIndex < 0 || nextIndex >= current.length) {
      return;
    }
    [current[index], current[nextIndex]] = [current[nextIndex], current[index]];
    this.isUpdating.set(true);
    this.api
      .reorderItemPhotos(
        this.itemId(),
        current.map((photo) => photo.id)
      )
      .subscribe({
        next: (response) => {
          this.photos.set(response.photos);
          this.isUpdating.set(false);
        },
        error: () => {
          this.isUpdating.set(false);
          this.loadError.set(true);
        }
      });
  }

  protected setPrimary(photo: ItemPhoto): void {
    this.isUpdating.set(true);
    this.api.setPrimaryItemPhoto(this.itemId(), photo.id).subscribe({
      next: (response) => {
        this.photos.set(response.photos);
        this.isUpdating.set(false);
      },
      error: () => {
        this.isUpdating.set(false);
        this.loadError.set(true);
      }
    });
  }

  protected removePhoto(photo: ItemPhoto): void {
    this.isUpdating.set(true);
    this.api.deleteItemPhoto(this.itemId(), photo.id).subscribe({
      next: () => {
        this.isUpdating.set(false);
        this.loadPhotos();
      },
      error: () => {
        this.isUpdating.set(false);
        this.loadError.set(true);
      }
    });
  }

  protected photoPreviewURL(photo: ItemPhoto): string {
    const url = photo.content_urls.thumbnail;
    if (url.startsWith('/api/')) {
      return url;
    }
    return url.startsWith('/') ? `/api${url}` : url;
  }

  private loadPhotos(): void {
    this.isLoading.set(true);
    this.loadError.set(false);
    this.api.listItemPhotos(this.itemId()).subscribe({
      next: (response) => {
        this.photos.set(response.photos);
        this.isLoading.set(false);
      },
      error: () => {
        this.photos.set([]);
        this.isLoading.set(false);
        this.loadError.set(true);
      }
    });
  }

  private uploadNext(files: File[], index: number): void {
    const file = files[index];
    if (!file) {
      this.isUploading.set(false);
      this.loadPhotos();
      return;
    }

    this.api.uploadItemPhoto(this.itemId(), file).subscribe({
      next: () => this.uploadNext(files, index + 1),
      error: () => {
        this.isUploading.set(false);
        this.uploadError.set(true);
      }
    });
  }
}
