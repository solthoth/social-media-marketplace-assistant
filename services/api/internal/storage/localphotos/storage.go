package localphotos

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
)

type Storage struct {
	root string
}

func NewStorage(root string) Storage {
	return Storage{root: root}
}

func (s Storage) Save(ctx context.Context, object photos.StorageObject, content io.Reader) (photos.StoredObject, error) {
	if err := ctx.Err(); err != nil {
		return photos.StoredObject{}, err
	}

	path := s.objectPath(object.StorageID, object.Variant, object.Extension)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return photos.StoredObject{}, err
	}

	file, err := os.Create(path)
	if err != nil {
		return photos.StoredObject{}, err
	}
	defer file.Close()

	written, err := io.Copy(file, content)
	if err != nil {
		return photos.StoredObject{}, err
	}
	return photos.StoredObject{StorageID: object.StorageID, SizeBytes: written}, nil
}

func (s Storage) Open(ctx context.Context, storageID string, variant domain.PhotoVariant) (io.ReadCloser, photos.ObjectInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	path, err := s.findVariantPath(storageID, variant)
	if err != nil {
		path, err = s.findVariantPath(storageID, domain.PhotoVariantOriginal)
		if err != nil {
			return nil, photos.ObjectInfo{}, err
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, photos.ObjectInfo{}, err
	}
	return file, photos.ObjectInfo{
		ContentType: contentTypeForExtension(filepath.Ext(path)),
		SizeBytes:   stat.Size(),
	}, nil
}

func (s Storage) Delete(ctx context.Context, storageID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	path := filepath.Join(s.root, filepath.Clean(storageID))
	if strings.HasPrefix(filepath.Clean(storageID), "..") {
		return photos.ErrInvalidPhoto
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func (s Storage) objectPath(storageID string, variant domain.PhotoVariant, extension string) string {
	return filepath.Join(s.root, filepath.Clean(storageID), string(variant)+extension)
}

func (s Storage) findVariantPath(storageID string, variant domain.PhotoVariant) (string, error) {
	if strings.HasPrefix(filepath.Clean(storageID), "..") {
		return "", photos.ErrInvalidPhoto
	}
	matches, err := filepath.Glob(filepath.Join(s.root, filepath.Clean(storageID), string(variant)+".*"))
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return "", photos.ErrPhotoNotFound
	}
	return matches[0], nil
}

func contentTypeForExtension(extension string) string {
	switch extension {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}
