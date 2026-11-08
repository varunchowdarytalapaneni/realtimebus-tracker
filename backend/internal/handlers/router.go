package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	v1 := r.Group("/api/v1")
	v1.POST("/telemetry", h.PostTelemetry)
	v1.POST("/telemetry/batch", h.PostTelemetryBatch)
	r.GET("/ws/track/:bus_id", h.TrackBus)

	return r
}
