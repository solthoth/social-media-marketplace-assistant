package httpserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/domain"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
)

type itemApplication interface {
	CreateItem(ctx context.Context, input items.CreateItemInput) (domain.Item, error)
	ListItems(ctx context.Context, filter items.ListItemsFilter) ([]domain.Item, error)
	GetItem(ctx context.Context, id string) (domain.Item, error)
	UpdateItem(ctx context.Context, id string, input items.UpdateItemInput) (domain.Item, error)
	ArchiveItem(ctx context.Context, id string) (domain.Item, error)
}

type itemRoutes struct {
	service itemApplication
}

type createItemRequest struct {
	Title                      string `json:"title"`
	Description                string `json:"description"`
	Category                   string `json:"category"`
	Size                       string `json:"size"`
	Condition                  string `json:"condition"`
	OriginalPurchasePriceCents int64  `json:"original_purchase_price_cents"`
	SellingPriceCents          int64  `json:"selling_price_cents"`
	Currency                   string `json:"currency"`
	Notes                      string `json:"notes"`
}

type updateItemRequest struct {
	Title                      *string `json:"title"`
	Description                *string `json:"description"`
	Category                   *string `json:"category"`
	Size                       *string `json:"size"`
	Condition                  *string `json:"condition"`
	OriginalPurchasePriceCents *int64  `json:"original_purchase_price_cents"`
	SellingPriceCents          *int64  `json:"selling_price_cents"`
	Currency                   *string `json:"currency"`
	Status                     *string `json:"status"`
	Notes                      *string `json:"notes"`
}

type itemResponse struct {
	ID                         string `json:"id"`
	Title                      string `json:"title"`
	Description                string `json:"description"`
	Category                   string `json:"category"`
	Size                       string `json:"size"`
	Condition                  string `json:"condition"`
	OriginalPurchasePriceCents int64  `json:"original_purchase_price_cents"`
	SellingPriceCents          int64  `json:"selling_price_cents"`
	Currency                   string `json:"currency"`
	Status                     string `json:"status"`
	Notes                      string `json:"notes"`
	CreatedAt                  string `json:"created_at"`
	UpdatedAt                  string `json:"updated_at"`
}

type listItemsResponse struct {
	Items []itemResponse `json:"items"`
}

func registerItemRoutes(router *gin.Engine, service itemApplication) {
	routes := itemRoutes{service: service}
	router.POST("/items", routes.create)
	router.GET("/items", routes.list)
	router.GET("/items/:id", routes.get)
	router.PATCH("/items/:id", routes.update)
	router.DELETE("/items/:id", routes.archive)
}

func (r itemRoutes) create(c *gin.Context) {
	var request createItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_json", "request body must be valid JSON"))
		return
	}

	item, err := r.service.CreateItem(c.Request.Context(), items.CreateItemInput{
		Title:                      request.Title,
		Description:                request.Description,
		Category:                   request.Category,
		Size:                       request.Size,
		Condition:                  request.Condition,
		OriginalPurchasePriceCents: request.OriginalPurchasePriceCents,
		SellingPriceCents:          request.SellingPriceCents,
		Currency:                   request.Currency,
		Notes:                      request.Notes,
	})
	if err != nil {
		writeItemServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, newItemResponse(item))
}

func (r itemRoutes) list(c *gin.Context) {
	filter := items.ListItemsFilter{}
	if statusValue := c.Query("status"); statusValue != "" {
		status := domain.InventoryStatus(statusValue)
		filter.Status = &status
	}

	result, err := r.service.ListItems(c.Request.Context(), filter)
	if err != nil {
		writeItemServiceError(c, err)
		return
	}

	response := listItemsResponse{Items: make([]itemResponse, 0, len(result))}
	for _, item := range result {
		response.Items = append(response.Items, newItemResponse(item))
	}
	c.JSON(http.StatusOK, response)
}

func (r itemRoutes) get(c *gin.Context) {
	item, err := r.service.GetItem(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeItemServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, newItemResponse(item))
}

func (r itemRoutes) update(c *gin.Context) {
	var request updateItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_json", "request body must be valid JSON"))
		return
	}

	input := items.UpdateItemInput{
		Title:                      request.Title,
		Description:                request.Description,
		Category:                   request.Category,
		Size:                       request.Size,
		Condition:                  request.Condition,
		OriginalPurchasePriceCents: request.OriginalPurchasePriceCents,
		SellingPriceCents:          request.SellingPriceCents,
		Currency:                   request.Currency,
		Notes:                      request.Notes,
	}
	if request.Status != nil {
		status := domain.InventoryStatus(*request.Status)
		input.Status = &status
	}

	item, err := r.service.UpdateItem(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		writeItemServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, newItemResponse(item))
}

func (r itemRoutes) archive(c *gin.Context) {
	_, err := r.service.ArchiveItem(c.Request.Context(), c.Param("id"))
	if err != nil {
		writeItemServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func newItemResponse(item domain.Item) itemResponse {
	return itemResponse{
		ID:                         item.ID,
		Title:                      item.Title,
		Description:                item.Description,
		Category:                   item.Category,
		Size:                       item.Size,
		Condition:                  item.Condition,
		OriginalPurchasePriceCents: item.OriginalPurchasePrice.AmountCents,
		SellingPriceCents:          item.SellingPrice.AmountCents,
		Currency:                   item.SellingPrice.Currency,
		Status:                     string(item.Status),
		Notes:                      item.Notes,
		CreatedAt:                  item.CreatedAt.Format(timeFormat),
		UpdatedAt:                  item.UpdatedAt.Format(timeFormat),
	}
}

func writeItemServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, items.ErrInvalidItem):
		writeError(c, NewAPIError(http.StatusBadRequest, "invalid_item", "item request is invalid"))
	case errors.Is(err, items.ErrItemNotFound):
		writeError(c, NewAPIError(http.StatusNotFound, "item_not_found", "item was not found"))
	default:
		writeError(c, NewAPIError(http.StatusInternalServerError, "internal_error", "request failed"))
	}
}
