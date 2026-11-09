package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"bus-tracker/internal/domain"
	"bus-tracker/internal/service"
)

func NewRouter(h *Handler, auth *service.AuthService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	v1 := r.Group("/api/v1")
	v1.POST("/auth/login", h.Login)
	driver := v1.Group("", RequireAuth(auth, domain.RoleDriver))
	driver.POST("/telemetry", h.PostTelemetry)
	driver.POST("/telemetry/batch", h.PostTelemetryBatch)
	r.GET("/ws/track/:bus_id", RequireAuth(auth), h.TrackBus)
	return r
}
