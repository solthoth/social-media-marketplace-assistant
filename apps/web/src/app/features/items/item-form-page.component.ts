import { NgIf } from '@angular/common';
import { Component, OnInit, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatButtonModule } from '@angular/material/button';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import {
  ApiClientService,
  InventoryItem,
  SaveInventoryItemRequest
} from '../../core/api-client.service';

@Component({
  selector: 'smm-item-form-page',
  standalone: true,
  imports: [
    MatButtonModule,
    MatFormFieldModule,
    MatInputModule,
    NgIf,
    ReactiveFormsModule,
    RouterLink
  ],
  template: `
    <section class="item-form-page">
      <div class="form-header">
        <div>
          <p class="eyebrow">Inventory</p>
          <h1>{{ isEditMode() ? 'Edit item' : 'New item' }}</h1>
          <p class="summary">
            Capture the details needed to keep this item ready for listing.
          </p>
        </div>
        <a matButton="outlined" class="secondary-action" routerLink="/items">
          Back to items
        </a>
      </div>

      <p *ngIf="loadError()" class="notice error" role="alert">
        Item could not be loaded.
      </p>

      <form
        class="item-form"
        [formGroup]="form"
        (ngSubmit)="saveDraft()"
        novalidate
      >
        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Title</mat-label>
          <input
            matInput
            data-testid="item-title"
            type="text"
            formControlName="title"
            autocomplete="off"
          />
          <mat-error *ngIf="showTitleError()" data-testid="title-error">
            Title is required.
          </mat-error>
        </mat-form-field>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Description</mat-label>
          <textarea
            matInput
            data-testid="item-description"
            formControlName="description"
            rows="4"
          ></textarea>
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Category</mat-label>
          <input
            matInput
            data-testid="item-category"
            type="text"
            formControlName="category"
            autocomplete="off"
          />
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Size</mat-label>
          <input
            matInput
            data-testid="item-size"
            type="text"
            formControlName="size"
            autocomplete="off"
          />
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Condition</mat-label>
          <input
            matInput
            data-testid="item-condition"
            type="text"
            formControlName="condition"
            autocomplete="off"
          />
        </mat-form-field>

        <div class="price-row">
          <mat-form-field appearance="outline">
            <mat-label>Original purchase price</mat-label>
            <input
              matInput
              data-testid="item-original-purchase-price"
              type="number"
              inputmode="decimal"
              min="0"
              step="0.01"
              formControlName="originalPurchasePrice"
            />
            <mat-error *ngIf="showOriginalPurchasePriceError()">
              Original purchase price must be zero or greater.
            </mat-error>
          </mat-form-field>

          <mat-form-field appearance="outline">
            <mat-label>Selling price</mat-label>
            <input
              matInput
              data-testid="item-selling-price"
              type="number"
              inputmode="decimal"
              min="0"
              step="0.01"
              formControlName="sellingPrice"
            />
            <mat-error *ngIf="showSellingPriceError()">
              Selling price must be zero or greater.
            </mat-error>
          </mat-form-field>

          <mat-form-field appearance="outline">
            <mat-label>Currency</mat-label>
            <input
              matInput
              data-testid="item-currency"
              type="text"
              maxlength="3"
              formControlName="currency"
              autocomplete="off"
            />
          </mat-form-field>
        </div>

        <mat-form-field appearance="outline" class="full-width">
          <mat-label>Notes</mat-label>
          <textarea
            matInput
            data-testid="item-notes"
            formControlName="notes"
            rows="3"
          ></textarea>
        </mat-form-field>

        <p *ngIf="saveError()" class="notice error" role="alert">
          Item could not be saved.
        </p>

        <div class="form-actions">
          <button
            matButton="filled"
            data-testid="save-draft"
            class="primary-action button-action"
            type="submit"
            [disabled]="isSaving() || isLoading()"
          >
            {{ isSaving() ? 'Saving...' : 'Save draft' }}
          </button>
          <a matButton="outlined" class="secondary-action" routerLink="/items">
            Cancel
          </a>
        </div>
      </form>
    </section>
  `
})
export class ItemFormPageComponent implements OnInit {
  private readonly api = inject(ApiClientService);
  private readonly fb = inject(FormBuilder);
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);

  protected readonly isLoading = signal(false);
  protected readonly isSaving = signal(false);
  protected readonly loadError = signal(false);
  protected readonly saveError = signal(false);
  protected readonly itemId = signal<string | null>(null);
  protected readonly isEditMode = signal(false);

  protected readonly form = this.fb.nonNullable.group({
    title: ['', Validators.required],
    description: [''],
    category: [''],
    size: [''],
    condition: [''],
    originalPurchasePrice: [0, Validators.min(0)],
    sellingPrice: [0, Validators.min(0)],
    currency: ['USD', [Validators.required, Validators.pattern(/[A-Za-z]{3}/)]],
    notes: ['']
  });

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (!id) {
      return;
    }

    this.itemId.set(id);
    this.isEditMode.set(true);
    this.loadItem(id);
  }

  protected saveDraft(): void {
    this.form.markAllAsTouched();
    this.saveError.set(false);

    if (this.form.invalid) {
      return;
    }

    this.isSaving.set(true);
    const payload = this.formPayload();
    const id = this.itemId();
    const request = id
      ? this.api.updateItem(id, payload)
      : this.api.createItem(payload);

    request.subscribe({
      next: () => {
        this.isSaving.set(false);
        void this.router.navigateByUrl('/items');
      },
      error: () => {
        this.isSaving.set(false);
        this.saveError.set(true);
      }
    });
  }

  protected showTitleError(): boolean {
    const control = this.form.controls.title;
    return control.invalid && control.touched;
  }

  protected showOriginalPurchasePriceError(): boolean {
    const control = this.form.controls.originalPurchasePrice;
    return control.invalid && control.touched;
  }

  protected showSellingPriceError(): boolean {
    const control = this.form.controls.sellingPrice;
    return control.invalid && control.touched;
  }

  private loadItem(id: string): void {
    this.isLoading.set(true);
    this.loadError.set(false);

    this.api.getItem(id).subscribe({
      next: (item) => {
        this.form.patchValue(this.formValue(item));
        this.isLoading.set(false);
      },
      error: () => {
        this.isLoading.set(false);
        this.loadError.set(true);
      }
    });
  }

  private formPayload(): SaveInventoryItemRequest {
    const value = this.form.getRawValue();
    return {
      title: value.title.trim(),
      description: value.description.trim(),
      category: value.category.trim(),
      size: value.size.trim(),
      condition: value.condition.trim(),
      original_purchase_price_cents: Math.round(
        Number(value.originalPurchasePrice || 0) * 100
      ),
      selling_price_cents: Math.round(Number(value.sellingPrice || 0) * 100),
      currency: value.currency.trim().toUpperCase(),
      notes: value.notes.trim()
    };
  }

  private formValue(item: InventoryItem) {
    return {
      title: item.title,
      description: item.description,
      category: item.category,
      size: item.size,
      condition: item.condition,
      originalPurchasePrice: item.original_purchase_price_cents / 100,
      sellingPrice: item.selling_price_cents / 100,
      currency: item.currency,
      notes: item.notes
    };
  }
}
