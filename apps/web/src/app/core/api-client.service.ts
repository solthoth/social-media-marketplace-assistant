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

export interface InventoryItem {
  id: string;
  title: string;
  description: string;
  category: string;
  size: string;
  condition: string;
  price_cents: number;
  currency: string;
  status: InventoryStatus;
  notes: string;
  created_at: string;
  updated_at: string;
}

export interface ListItemsResponse {
  items: InventoryItem[];
}

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
}
