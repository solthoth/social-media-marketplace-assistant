package items

import (
	"context"
	"slices"
	"sync"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
)

type memoryRepository struct {
	mu    sync.RWMutex
	items map[string]domain.Item
}

func newMemoryRepository() *memoryRepository {
	return &memoryRepository{
		items: map[string]domain.Item{},
	}
}

func (r *memoryRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[item.ID] = item
	return item, nil
}

func (r *memoryRepository) List(ctx context.Context, filter ListItemsFilter) ([]domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	items := make([]domain.Item, 0, len(r.items))
	for _, item := range r.items {
		if filter.Status != nil && item.Status != *filter.Status {
			continue
		}
		items = append(items, item)
	}
	slices.SortFunc(items, func(a domain.Item, b domain.Item) int {
		return a.CreatedAt.Compare(b.CreatedAt)
	})
	return items, nil
}

func (r *memoryRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return domain.Item{}, ErrItemNotFound
	}
	return item, nil
}

func (r *memoryRepository) Update(ctx context.Context, item domain.Item) (domain.Item, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[item.ID]; !ok {
		return domain.Item{}, ErrItemNotFound
	}
	r.items[item.ID] = item
	return item, nil
}
