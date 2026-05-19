package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/items"
	"github.com/solthoth/social-media-marketplace-assistant/services/api/internal/photos"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

const timeFormat = time.RFC3339

type RouterDependencies struct {
	ItemService  *items.Service
	PhotoService *photos.Service
}

func NewRouter(dependencies ...RouterDependencies) http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", healthz)
	router.GET("/", root)
	registerSwaggerRoutes(router)
	if len(dependencies) > 0 && dependencies[0].ItemService != nil {
		registerItemRoutes(router, dependencies[0].ItemService)
	}
	if len(dependencies) > 0 && dependencies[0].PhotoService != nil {
		registerPhotoRoutes(router, dependencies[0].PhotoService)
	}

	return router
}

func root(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"service": "social-media-marketplace-assistant-api",
		"status":  "ok",
	})
}

func healthz(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{
		Status:  "ok",
		Service: "social-media-marketplace-assistant-api",
		Time:    time.Now().UTC().Format(timeFormat),
	})
}
