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

	directory, err := storageRelativeDirectory(object.StorageID)
	if err != nil {
		return photos.StoredObject{}, err
	}
	filename, err := variantFilename(object.Variant, object.Extension)
	if err != nil {
		return photos.StoredObject{}, err
	}
	root, err := s.openRoot()
	if err != nil {
		return photos.StoredObject{}, err
	}
	defer root.Close()

	if err := root.MkdirAll(directory, 0o750); err != nil {
		return photos.StoredObject{}, err
	}

	file, err := root.Create(filepath.Join(directory, filename))
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
	directory, err := storageRelativeDirectory(storageID)
	if err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	root, err := s.openRoot()
	if err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	defer root.Close()

	filename, err := findVariantFilename(root, directory, variant)
	if err != nil {
		filename, err = findVariantFilename(root, directory, domain.PhotoVariantOriginal)
		if err != nil {
			return nil, photos.ObjectInfo{}, err
		}
	}

	file, err := root.Open(filepath.Join(directory, filename))
	if err != nil {
		return nil, photos.ObjectInfo{}, err
	}
	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, photos.ObjectInfo{}, err
	}
	return file, photos.ObjectInfo{
		ContentType: contentTypeForExtension(filepath.Ext(filename)),
		SizeBytes:   stat.Size(),
	}, nil
}

func (s Storage) Delete(ctx context.Context, storageID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	directory, err := storageRelativeDirectory(storageID)
	if err != nil {
		return err
	}
	root, err := s.openRoot()
	if err != nil {
		return err
	}
	defer root.Close()

	if err := root.RemoveAll(directory); err != nil {
		return err
	}
	return nil
}

func (s Storage) openRoot() (*os.Root, error) {
	if err := os.MkdirAll(s.root, 0o750); err != nil {
		return nil, err
	}
	return os.OpenRoot(s.root)
}

func storageRelativeDirectory(storageID string) (string, error) {
	if !isSafeStorageID(storageID) {
		return "", photos.ErrInvalidPhoto
	}
	return filepath.Join(strings.Split(storageID, "/")...), nil
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

func findVariantFilename(root *os.Root, directory string, variant domain.PhotoVariant) (string, error) {
	if !variant.IsValid() {
		return "", photos.ErrInvalidPhoto
	}
	for _, extension := range []string{".jpg", ".png", ".webp"} {
		filename := string(variant) + extension
		if _, err := root.Stat(filepath.Join(directory, filename)); err == nil {
			return filename, nil
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
