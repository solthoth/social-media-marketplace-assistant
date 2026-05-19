import { InventoryStatus } from './api-client.service';

export const inventoryStatusLabels: Record<InventoryStatus, string> = {
  draft: 'Draft',
  ready_to_list: 'Ready to list',
  listed: 'Listed',
  sold: 'Sold',
  archived: 'Archived'
};

const inventoryStatuses: InventoryStatus[] = [
  'draft',
  'ready_to_list',
  'listed',
  'sold',
  'archived'
];

const transitionTargets: Record<InventoryStatus, InventoryStatus[]> = {
  draft: ['ready_to_list', 'archived'],
  ready_to_list: ['draft', 'listed', 'archived'],
  listed: ['ready_to_list', 'sold', 'archived'],
  sold: ['listed', 'archived'],
  archived: ['draft']
};

export function inventoryStatusOptions(
  current: InventoryStatus
): InventoryStatus[] {
  return inventoryStatuses.filter(
    (status) =>
      status === current || transitionTargets[current].includes(status)
  );
}
