import { Routes } from '@angular/router';
import { DashboardComponent } from './features/dashboard/dashboard.component';
import { InventoryPageComponent } from './features/inventory/inventory-page.component';
import { ItemFormPageComponent } from './features/items/item-form-page.component';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    component: DashboardComponent
  },
  {
    path: 'items',
    component: InventoryPageComponent
  },
  {
    path: 'items/new',
    component: ItemFormPageComponent
  },
  {
    path: 'items/:id/edit',
    component: ItemFormPageComponent
  }
];
