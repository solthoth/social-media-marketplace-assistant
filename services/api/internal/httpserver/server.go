package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type healthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Time    string `json:"time"`
}

func NewRouter() http.Handler {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET("/healthz", healthz)
	router.GET("/", root)

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
		Time:    time.Now().UTC().Format(time.RFC3339),
	})
}
