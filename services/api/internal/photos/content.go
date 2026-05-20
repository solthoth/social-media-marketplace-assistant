package photos

import (
	"context"
	"io"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
)

func (s Service) OpenPhotoContent(ctx context.Context, photo domain.ItemPhoto) ([]byte, error) {
	content, _, err := s.storage.Open(ctx, photo.StorageID, domain.PhotoVariantOriginal)
	if err != nil {
		return nil, err
	}
	defer content.Close()
	return io.ReadAll(content)
}
