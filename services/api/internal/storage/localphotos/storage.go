package localphotos

import (
	"context"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
)

var safeStorageIDPattern = regexp.MustCompile(`\A[A-Za-z0-9_-]+(/[A-Za-z0-9_-]+)*\z`)

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

	directory, err := s.storageDirectory(object.StorageID)
	if err != nil {
		return photos.StoredObject{}, err
	}
	filename, err := variantFilename(object.Variant, object.Extension)
	if err != nil {
		return photos.StoredObject{}, err
	}
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return photos.StoredObject{}, err
	}

	file, err := os.Create(filepath.Join(directory, filename))
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
	directory, err := s.storageDirectory(storageID)
	if err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	path, err := findVariantPath(directory, variant)
	if err != nil {
		path, err = findVariantPath(directory, domain.PhotoVariantOriginal)
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
	directory, err := s.storageDirectory(storageID)
	if err != nil {
		return err
	}
	if err := os.RemoveAll(directory); err != nil {
		return err
	}
	return nil
}

func (s Storage) storageDirectory(storageID string) (string, error) {
	if !isSafeStorageID(storageID) {
		return "", photos.ErrInvalidPhoto
	}

	root, err := filepath.Abs(s.root)
	if err != nil {
		return "", err
	}
	parts := strings.Split(storageID, "/")
	directory := filepath.Join(append([]string{root}, parts...)...)
	relative, err := filepath.Rel(root, directory)
	if err != nil {
		return "", err
	}
	if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", photos.ErrInvalidPhoto
	}
	return directory, nil
}

func isSafeStorageID(storageID string) bool {
	if storageID == "" || strings.Contains(storageID, "\\") || strings.ContainsRune(storageID, 0) {
		return false
	}
	if path.Clean(storageID) != storageID || path.IsAbs(storageID) || filepath.IsAbs(storageID) {
		return false
	}
	return safeStorageIDPattern.MatchString(storageID)
}

func variantFilename(variant domain.PhotoVariant, extension string) (string, error) {
	if !variant.IsValid() {
		return "", photos.ErrInvalidPhoto
	}
	if !isSafeExtension(extension) {
		return "", photos.ErrInvalidPhoto
	}
	return string(variant) + extension, nil
}

func findVariantPath(directory string, variant domain.PhotoVariant) (string, error) {
	if !variant.IsValid() {
		return "", photos.ErrInvalidPhoto
	}
	for _, extension := range []string{".jpg", ".png", ".webp"} {
		candidate := filepath.Join(directory, string(variant)+extension)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", photos.ErrPhotoNotFound
}

func isSafeExtension(extension string) bool {
	switch extension {
	case ".jpg", ".png", ".webp":
		return true
	default:
		return false
	}
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
