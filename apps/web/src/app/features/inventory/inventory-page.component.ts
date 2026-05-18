import { CurrencyPipe, DatePipe, NgFor, NgIf } from '@angular/common';
import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { MatButtonModule } from '@angular/material/button';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatSelectModule } from '@angular/material/select';
import { RouterLink } from '@angular/router';
import {
  ApiClientService,
  InventoryItem,
  InventoryStatus
} from '../../core/api-client.service';

type InventoryStatusFilter = InventoryStatus | 'all';

const statusLabels: Record<InventoryStatus, string> = {
  draft: 'Draft',
  ready_to_list: 'Ready to list',
  listed: 'Listed',
  sold: 'Sold',
  archived: 'Archived'
};

@Component({
  selector: 'smm-inventory-page',
  standalone: true,
  imports: [
    CurrencyPipe,
    DatePipe,
    MatButtonModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatSelectModule,
    NgFor,
    NgIf,
    RouterLink
  ],
  template: `
    <section class="inventory-page">
      <div class="inventory-header">
        <div>
          <p class="eyebrow">Inventory</p>
          <h1>Items</h1>
          <p class="summary">
            Review active inventory, find items quickly, and filter by listing
            workflow status.
          </p>
        </div>
        <div class="inventory-count" aria-live="polite">
          {{ filteredItems().length }}
          {{ filteredItems().length === 1 ? 'item' : 'items' }}
        </div>
      </div>

      <div class="inventory-toolbar" aria-label="Inventory filters">
        <mat-form-field appearance="outline">
          <mat-label>Search</mat-label>
          <input
            matInput
            data-testid="inventory-search"
            type="search"
            [value]="searchTerm()"
            (input)="setSearch($event)"
            placeholder="Title, category, or notes"
          />
        </mat-form-field>

        <mat-form-field appearance="outline">
          <mat-label>Status</mat-label>
          <select
            matNativeControl
            data-testid="status-filter"
            [value]="statusFilter()"
            (change)="setStatus($event)"
          >
            <option value="all">All statuses</option>
            <option value="draft">Draft</option>
            <option value="ready_to_list">Ready to list</option>
            <option value="listed">Listed</option>
            <option value="sold">Sold</option>
            <option value="archived">Archived</option>
          </select>
        </mat-form-field>
      </div>

      <p *ngIf="loadError()" class="notice error" role="alert">
        Inventory could not be loaded.
      </p>

      <p *ngIf="!loadError() && isLoading()" class="notice">Loading items...</p>

      <section
        *ngIf="!loadError() && !isLoading() && filteredItems().length > 0"
        class="inventory-grid"
        aria-label="Inventory items"
      >
        <mat-card
          *ngFor="let item of filteredItems()"
          appearance="outlined"
          class="inventory-card"
        >
          <mat-card-header class="card-heading">
            <mat-card-title>{{ item.title }}</mat-card-title>
            <span class="status-badge">{{ statusLabel(item.status) }}</span>
          </mat-card-header>

          <mat-card-content>
            <dl class="item-facts">
              <div>
                <dt>Category</dt>
                <dd>{{ item.category || 'Uncategorized' }}</dd>
              </div>
              <div>
                <dt>Price</dt>
                <dd>{{ item.price_cents / 100 | currency: item.currency }}</dd>
              </div>
              <div>
                <dt>Size</dt>
                <dd>{{ item.size || 'Not set' }}</dd>
              </div>
              <div>
                <dt>Updated</dt>
                <dd>{{ item.updated_at | date: 'mediumDate' }}</dd>
              </div>
            </dl>

            <p class="item-description" *ngIf="item.description">
              {{ item.description }}
            </p>
            <p class="item-notes" *ngIf="item.notes">{{ item.notes }}</p>
          </mat-card-content>

          <mat-card-actions align="end">
            <a
              matButton="outlined"
              class="secondary-action"
              [routerLink]="['/items', item.id, 'edit']"
            >
              Edit
            </a>
          </mat-card-actions>
        </mat-card>
      </section>

      <section
        *ngIf="!loadError() && !isLoading() && filteredItems().length === 0"
        class="empty-state"
      >
        <h2>No inventory items</h2>
        <p>Items that match the current filters will appear here.</p>
      </section>
    </section>
  `
})
export class InventoryPageComponent implements OnInit {
  private readonly api = inject(ApiClientService);

  protected readonly items = signal<InventoryItem[]>([]);
  protected readonly isLoading = signal(true);
  protected readonly loadError = signal(false);
  protected readonly searchTerm = signal('');
  protected readonly statusFilter = signal<InventoryStatusFilter>('all');
  protected readonly filteredItems = computed(() => {
    const query = this.searchTerm().trim().toLowerCase();
    if (!query) {
      return this.items();
    }

    return this.items().filter((item) => {
      const searchable = [item.title, item.category, item.notes]
        .join(' ')
        .toLowerCase();
      return searchable.includes(query);
    });
  });

  ngOnInit(): void {
    this.loadItems();
  }

  protected setSearch(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.searchTerm.set(input.value);
  }

  protected setStatus(event: Event): void {
    const select = event.target as HTMLSelectElement;
    const status = select.value as InventoryStatusFilter;
    this.statusFilter.set(status);
    this.loadItems();
  }

  protected statusLabel(status: InventoryStatus): string {
    return statusLabels[status];
  }

  private loadItems(): void {
    this.isLoading.set(true);
    this.loadError.set(false);
    const status = this.statusFilter();

    this.api.listItems(status === 'all' ? undefined : status).subscribe({
      next: (response) => {
        this.items.set(response.items);
        this.isLoading.set(false);
      },
      error: () => {
        this.items.set([]);
        this.isLoading.set(false);
        this.loadError.set(true);
      }
    });
  }
}
