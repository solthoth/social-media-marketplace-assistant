package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"slices"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
)

type PhotoRepository struct {
	db *sql.DB
}

func NewPhotoRepository(db *sql.DB) PhotoRepository {
	return PhotoRepository{db: db}
}

func (r PhotoRepository) CreatePhoto(ctx context.Context, photo domain.ItemPhoto) (domain.ItemPhoto, error) {
	_, err := r.db.ExecContext(ctx, `
insert into item_photos (
  id, item_id, storage_id, filename, mime_type, sort_order, is_primary, created_at
) values (?, ?, ?, ?, ?, ?, ?, ?)`,
		photo.ID,
		photo.ItemID,
		photo.StorageID,
		photo.Filename,
		photo.MimeType,
		photo.SortOrder,
		boolToInt(photo.IsPrimary),
		formatTime(photo.CreatedAt),
	)
	if err != nil {
		return domain.ItemPhoto{}, err
	}
	return photo, nil
}

func (r PhotoRepository) ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error) {
	rows, err := r.db.QueryContext(ctx, `
select id, item_id, storage_id, filename, mime_type, sort_order, is_primary, created_at
from item_photos
where item_id = ?
order by sort_order asc, created_at asc`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.ItemPhoto{}
	for rows.Next() {
		photo, err := scanPhoto(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, photo)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r PhotoRepository) GetPhoto(ctx context.Context, itemID string, photoID string) (domain.ItemPhoto, error) {
	row := r.db.QueryRowContext(ctx, `
select id, item_id, storage_id, filename, mime_type, sort_order, is_primary, created_at
from item_photos
where item_id = ? and id = ?`, itemID, photoID)

	photo, err := scanPhoto(row)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ItemPhoto{}, photos.ErrPhotoNotFound
	}
	if err != nil {
		return domain.ItemPhoto{}, err
	}
	return photo, nil
}

func (r PhotoRepository) DeletePhoto(ctx context.Context, itemID string, photoID string) error {
	result, err := r.db.ExecContext(ctx, "delete from item_photos where item_id = ? and id = ?", itemID, photoID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return photos.ErrPhotoNotFound
	}
	_, err = r.db.ExecContext(ctx, `
update item_photos
set is_primary = 1
where id = (
  select id from item_photos
  where item_id = ?
  order by sort_order asc, created_at asc
  limit 1
) and not exists (
  select 1 from item_photos where item_id = ? and is_primary = 1
)`, itemID, itemID)
	return err
}

func (r PhotoRepository) ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error) {
	existing, err := r.ListPhotos(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if !samePhotoSet(existing, photoIDs) {
		return nil, photos.ErrInvalidPhoto
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for index, photoID := range photoIDs {
		if _, err := tx.ExecContext(ctx, "update item_photos set sort_order = ? where item_id = ? and id = ?", index, itemID, photoID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.ListPhotos(ctx, itemID)
}

type photoScanner interface {
	Scan(dest ...any) error
}

func scanPhoto(scanner photoScanner) (domain.ItemPhoto, error) {
	var photo domain.ItemPhoto
	var isPrimary int
	var createdAt string
	err := scanner.Scan(
		&photo.ID,
		&photo.ItemID,
		&photo.StorageID,
		&photo.Filename,
		&photo.MimeType,
		&photo.SortOrder,
		&isPrimary,
		&createdAt,
	)
	if err != nil {
		return domain.ItemPhoto{}, err
	}
	photo.IsPrimary = isPrimary == 1
	photo.CreatedAt, err = parseTime(createdAt)
	if err != nil {
		return domain.ItemPhoto{}, err
	}
	return photo, nil
}

func samePhotoSet(existing []domain.ItemPhoto, photoIDs []string) bool {
	if len(existing) != len(photoIDs) {
		return false
	}
	existingIDs := make([]string, 0, len(existing))
	for _, photo := range existing {
		existingIDs = append(existingIDs, photo.ID)
	}
	slices.Sort(existingIDs)
	requested := slices.Clone(photoIDs)
	slices.Sort(requested)
	return slices.Equal(existingIDs, requested)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
