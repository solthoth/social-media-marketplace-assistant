package httpserver

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
)

type photoApplication interface {
	UploadPhoto(ctx context.Context, itemID string, input photos.UploadPhotoInput) (domain.ItemPhoto, error)
	ListPhotos(ctx context.Context, itemID string) ([]domain.ItemPhoto, error)
	OpenPhoto(ctx context.Context, itemID string, photoID string, variant domain.PhotoVariant) (io.ReadCloser, photos.ObjectInfo, error)
	DeletePhoto(ctx context.Context, itemID string, photoID string) error
	ReorderPhotos(ctx context.Context, itemID string, photoIDs []string) ([]domain.ItemPhoto, error)
}

type photoRoutes struct {
	service photoApplication
}

type photoResponse struct {
	ID          string            `json:"id"`
	ItemID      string            `json:"item_id"`
	Filename    string            `json:"filename"`
	MimeType    string            `json:"mime_type"`
	SortOrder   int               `json:"sort_order"`
	IsPrimary   bool              `json:"is_primary"`
	ContentURLs map[string]string `json:"content_urls"`
	CreatedAt   string            `json:"created_at"`
}

type listPhotosResponse struct {
	Photos []photoResponse `json:"photos"`
}

type reorderPhotosRequest struct {
	PhotoIDs []string `json:"photo_ids"`
}

func registerPhotoRoutes(router *gin.Engine, service photoApplication) {
	routes := photoRoutes{service: service}
	router.POST("/items/:id/photos", routes.upload)
	router.GET("/items/:id/photos", routes.list)
	router.GET("/items/:id/photos/:photoID/content", routes.content)
	router.DELETE("/items/:id/photos/:photoID", routes.delete)
	router.PATCH("/items/:id/photos/order", routes.reorder)
}

func (r photoRoutes) upload(c *gin.Context) {
	file, err := c.FormFile("photo")
	if err != nil {
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_photo", "photo upload is invalid"))
		return
	}
	content, err := file.Open()
	if err != nil {
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_photo", "photo upload is invalid"))
		return
	}
	defer content.Close()

	photo, err := r.service.UploadPhoto(c.Request.Context(), c.Param("id"), photos.UploadPhotoInput{
		Filename: file.Filename,
		Content:  content,
	})
	if err != nil {
		writePhotoServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, newPhotoResponse(c, photo))
}

func (r photoRoutes) list(c *gin.Context) {
	result, err := r.service.ListPhotos(c.Request.Context(), c.Param("id"))
	if err != nil {
		writePhotoServiceError(c, err)
		return
	}

	response := listPhotosResponse{Photos: make([]photoResponse, 0, len(result))}
	for _, photo := range result {
		response.Photos = append(response.Photos, newPhotoResponse(c, photo))
	}
	c.JSON(http.StatusOK, response)
}

func (r photoRoutes) content(c *gin.Context) {
	variant := domain.PhotoVariant(c.DefaultQuery("variant", string(domain.PhotoVariantOriginal)))
	content, info, err := r.service.OpenPhoto(c.Request.Context(), c.Param("id"), c.Param("photoID"), variant)
	if err != nil {
		writePhotoServiceError(c, err)
		return
	}
	defer content.Close()

	c.DataFromReader(http.StatusOK, info.SizeBytes, info.ContentType, content, nil)
}

func (r photoRoutes) delete(c *gin.Context) {
	if err := r.service.DeletePhoto(c.Request.Context(), c.Param("id"), c.Param("photoID")); err != nil {
		writePhotoServiceError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (r photoRoutes) reorder(c *gin.Context) {
	var request reorderPhotosRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_json", "request body must be valid JSON"))
		return
	}

	result, err := r.service.ReorderPhotos(c.Request.Context(), c.Param("id"), request.PhotoIDs)
	if err != nil {
		writePhotoServiceError(c, err)
		return
	}
	response := listPhotosResponse{Photos: make([]photoResponse, 0, len(result))}
	for _, photo := range result {
		response.Photos = append(response.Photos, newPhotoResponse(c, photo))
	}
	c.JSON(http.StatusOK, response)
}

func newPhotoResponse(c *gin.Context, photo domain.ItemPhoto) photoResponse {
	return photoResponse{
		ID:        photo.ID,
		ItemID:    photo.ItemID,
		Filename:  photo.Filename,
		MimeType:  photo.MimeType,
		SortOrder: photo.SortOrder,
		IsPrimary: photo.IsPrimary,
		ContentURLs: map[string]string{
			string(domain.PhotoVariantOriginal):  photoContentURL(c, photo, domain.PhotoVariantOriginal),
			string(domain.PhotoVariantMedium):    photoContentURL(c, photo, domain.PhotoVariantMedium),
			string(domain.PhotoVariantThumbnail): photoContentURL(c, photo, domain.PhotoVariantThumbnail),
		},
		CreatedAt: photo.CreatedAt.Format(timeFormat),
	}
}

func photoContentURL(c *gin.Context, photo domain.ItemPhoto, variant domain.PhotoVariant) string {
	return "/items/" + photo.ItemID + "/photos/" + photo.ID + "/content?variant=" + string(variant)
}

func writePhotoServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, photos.ErrInvalidPhoto):
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_photo", "photo request is invalid"))
	case errors.Is(err, photos.ErrPhotoNotFound):
		writeError(c, NewAPIError(http.StatusNotFound, "photo_not_found", "photo was not found"))
	case errors.Is(err, items.ErrItemNotFound):
		writeError(c, NewAPIError(http.StatusNotFound, "item_not_found", "item was not found"))
	default:
		writeError(c, NewAPIError(http.StatusInternalServerError, "internal_error", "request failed"))
	}
}
