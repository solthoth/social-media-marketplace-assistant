package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
)

type ItemRepository struct {
	db *sql.DB
}

func NewItemRepository(db *sql.DB) ItemRepository {
	return ItemRepository{
		db: db,
	}
}

func (r ItemRepository) Create(ctx context.Context, item domain.Item) (domain.Item, error) {
	_, err := r.db.ExecContext(ctx, `
insert into items (
  id, title, description, category, size, condition, original_purchase_price_cents, selling_price_cents, currency, status, notes, created_at, updated_at
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		item.ID,
		item.Title,
		item.Description,
		item.Category,
		item.Size,
		item.Condition,
		item.OriginalPurchasePrice.AmountCents,
		item.SellingPrice.AmountCents,
		item.SellingPrice.Currency,
		item.Status,
		item.Notes,
		formatTime(item.CreatedAt),
		formatTime(item.UpdatedAt),
	)
	if err != nil {
		return domain.Item{}, err
	}
	return item, nil
}

func (r ItemRepository) List(ctx context.Context, filter items.ListItemsFilter) ([]domain.Item, error) {
	query := `
select id, title, description, category, size, condition, original_purchase_price_cents, selling_price_cents, currency, status, notes, created_at, updated_at
from items`
	args := []any{}
	if filter.Status != nil {
		query += " where status = ?"
		args = append(args, string(*filter.Status))
	}
	query += " order by created_at desc"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.Item{}
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r ItemRepository) Get(ctx context.Context, id string) (domain.Item, error) {
	row := r.db.QueryRowContext(ctx, `
select id, title, description, category, size, condition, original_purchase_price_cents, selling_price_cents, currency, status, notes, created_at, updated_at
from items
where id = ?`, id)

	item, err := scanItem(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Item{}, items.ErrItemNotFound
	}
	if err != nil {
		return domain.Item{}, err
	}
	return item, nil
}

func (r ItemRepository) Update(ctx context.Context, item domain.Item) (domain.Item, error) {
	result, err := r.db.ExecContext(ctx, `
update items
set title = ?,
    description = ?,
    category = ?,
    size = ?,
    condition = ?,
    original_purchase_price_cents = ?,
    selling_price_cents = ?,
    currency = ?,
    status = ?,
    notes = ?,
    updated_at = ?
where id = ?`,
		item.Title,
		item.Description,
		item.Category,
		item.Size,
		item.Condition,
		item.OriginalPurchasePrice.AmountCents,
		item.SellingPrice.AmountCents,
		item.SellingPrice.Currency,
		item.Status,
		item.Notes,
		formatTime(item.UpdatedAt),
		item.ID,
	)
	if err != nil {
		return domain.Item{}, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return domain.Item{}, err
	}
	if rowsAffected == 0 {
		return domain.Item{}, items.ErrItemNotFound
	}
	return item, nil
}

type itemScanner interface {
	Scan(dest ...any) error
}

func scanItem(scanner itemScanner) (domain.Item, error) {
	var item domain.Item
	var status string
	var createdAt string
	var updatedAt string
	err := scanner.Scan(
		&item.ID,
		&item.Title,
		&item.Description,
		&item.Category,
		&item.Size,
		&item.Condition,
		&item.OriginalPurchasePrice.AmountCents,
		&item.SellingPrice.AmountCents,
		&item.SellingPrice.Currency,
		&status,
		&item.Notes,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return domain.Item{}, err
	}

	item.OriginalPurchasePrice.Currency = item.SellingPrice.Currency
	item.Status = domain.InventoryStatus(status)
	item.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return domain.Item{}, err
	}
	item.UpdatedAt, err = parseTime(updatedAt)
	if err != nil {
		return domain.Item{}, err
	}
	return item, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, value)
}
