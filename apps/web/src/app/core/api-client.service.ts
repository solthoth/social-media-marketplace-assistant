import { HttpClient } from '@angular/common/http';
import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';

export interface HealthResponse {
  status: string;
  service: string;
  time: string;
}

export type InventoryStatus =
  | 'draft'
  | 'ready_to_list'
  | 'listed'
  | 'sold'
  | 'archived';

export type Currency = 'USD';

export type PhotoVariant = 'original' | 'medium' | 'thumbnail';

export interface InventoryItem {
  id: string;
  title: string;
  description: string;
  category: string;
  size: string;
  condition: string;
  original_purchase_price_cents: number;
  selling_price_cents: number;
  currency: Currency;
  status: InventoryStatus;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface ListItemsResponse {
  items: InventoryItem[];
}

export interface ItemPhoto {
  id: string;
  item_id: string;
  filename: string;
  mime_type: string;
  sort_order: number;
  is_primary: boolean;
  content_urls: Record<PhotoVariant, string>;
  created_at: string;
}

export interface ListItemPhotosResponse {
  photos: ItemPhoto[];
}

export type EnrichmentJobStatus =
  | 'queued'
  | 'processing'
  | 'completed'
  | 'failed';

export interface ItemDetailSuggestion {
  description: string;
  category: string;
  size: string;
  condition: string;
  notes: string;
}

export interface EnrichmentJob {
  id: string;
  item_id: string;
  status: EnrichmentJobStatus;
  provider: string;
  model: string;
  requested_at: string;
  started_at: string | null;
  completed_at: string | null;
  applied_at: string | null;
  error_message: string;
  suggestion: ItemDetailSuggestion;
}

export interface ApplyEnrichmentJobResponse {
  item: InventoryItem;
  job: EnrichmentJob;
  applied_fields: string[];
}

export interface SaveInventoryItemRequest {
  title: string;
  description: string;
  category: string;
  size: string;
  condition: string;
  original_purchase_price_cents: number;
  selling_price_cents: number;
  currency: Currency;
  notes: string;
}

export type UpdateInventoryItemRequest = Partial<SaveInventoryItemRequest> & {
  status?: InventoryStatus;
};

@Injectable({
  providedIn: 'root'
})
export class ApiClientService {
  private readonly http = inject(HttpClient);

  health(): Observable<HealthResponse> {
    return this.http.get<HealthResponse>('/api/healthz');
  }

  listItems(status?: InventoryStatus): Observable<ListItemsResponse> {
    return this.http.get<ListItemsResponse>('/api/items', {
      params: status ? { status } : {}
    });
  }

  getItem(id: string): Observable<InventoryItem> {
    return this.http.get<InventoryItem>(`/api/items/${id}`);
  }

  createItem(request: SaveInventoryItemRequest): Observable<InventoryItem> {
    return this.http.post<InventoryItem>('/api/items', request);
  }

  updateItem(
    id: string,
    request: UpdateInventoryItemRequest
  ): Observable<InventoryItem> {
    return this.http.patch<InventoryItem>(`/api/items/${id}`, request);
  }

  listItemPhotos(itemId: string): Observable<ListItemPhotosResponse> {
    return this.http.get<ListItemPhotosResponse>(`/api/items/${itemId}/photos`);
  }

  uploadItemPhoto(itemId: string, file: File): Observable<ItemPhoto> {
    const body = new FormData();
    body.append('photo', file);
    return this.http.post<ItemPhoto>(`/api/items/${itemId}/photos`, body);
  }

  reorderItemPhotos(
    itemId: string,
    photoIds: string[]
  ): Observable<ListItemPhotosResponse> {
    return this.http.patch<ListItemPhotosResponse>(
      `/api/items/${itemId}/photos/order`,
      {
        photo_ids: photoIds
      }
    );
  }

  setPrimaryItemPhoto(
    itemId: string,
    photoId: string
  ): Observable<ListItemPhotosResponse> {
    return this.http.patch<ListItemPhotosResponse>(
      `/api/items/${itemId}/photos/${photoId}/primary`,
      {}
    );
  }

  deleteItemPhoto(itemId: string, photoId: string): Observable<void> {
    return this.http.delete<void>(`/api/items/${itemId}/photos/${photoId}`);
  }

  createItemEnrichmentJob(itemId: string): Observable<EnrichmentJob> {
    return this.http.post<EnrichmentJob>(
      `/api/items/${itemId}/enrichment-jobs`,
      {}
    );
  }

  getItemEnrichmentJob(
    itemId: string,
    jobId: string
  ): Observable<EnrichmentJob> {
    return this.http.get<EnrichmentJob>(
      `/api/items/${itemId}/enrichment-jobs/${jobId}`
    );
  }

  applyItemEnrichmentJob(
    itemId: string,
    jobId: string
  ): Observable<ApplyEnrichmentJobResponse> {
    return this.http.post<ApplyEnrichmentJobResponse>(
      `/api/items/${itemId}/enrichment-jobs/${jobId}/apply`,
      {}
    );
  }
}
