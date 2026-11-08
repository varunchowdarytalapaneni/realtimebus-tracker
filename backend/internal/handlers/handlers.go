package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"bus-tracker/internal/domain"
	"bus-tracker/internal/hub"
)

const (
	writeWait = 10 * time.Second
	pongWait  = 60 * time.Second
	pingEvery = 30 * time.Second
)

// TODO(auth): restrict origins and require a token before production.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Handler struct {
	svc *TelemetryService
	hub *hub.Hub
}

func NewHandler(svc *TelemetryService, h *hub.Hub) *Handler {
	return &Handler{svc: svc, hub: h}
}

// POST /api/v1/telemetry
func (h *Handler) PostTelemetry(c *gin.Context) {
	var t domain.Telemetry
	if err := c.ShouldBindJSON(&t); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.ingest(c, []domain.Telemetry{t})
}

// POST /api/v1/telemetry/batch  (used by the Flutter app to flush its offline buffer)
func (h *Handler) PostTelemetryBatch(c *gin.Context) {
	var body struct {
		Points []domain.Telemetry `json:"points" binding:"required,min=1,max=500"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.ingest(c, body.Points)
}

func (h *Handler) ingest(c *gin.Context, points []domain.Telemetry) {
	err := h.svc.Ingest(c.Request.Context(), points)
	switch {
	case errors.Is(err, domain.ErrInvalidTelemetry):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case err != nil:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	default:
		c.JSON(http.StatusAccepted, gin.H{"accepted": len(points)})
	}
}

// GET /ws/track/:bus_id
func (h *Handler) TrackBus(c *gin.Context) {
	busID, err := strconv.ParseInt(c.Param("bus_id"), 10, 64)
	if err != nil || busID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bus_id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client := &hub.Client{BusID: busID, Send: make(chan []byte, 16)}
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	// Immediately send the last known position so the map isn't empty until the next ping.
	// (Clients should ignore messages older than one they've already shown, via "timestamp".)
	if last, err := h.svc.Latest(c.Request.Context(), busID); err == nil && last != nil {
		if data, err := json.Marshal(last); err == nil {
			client.Send <- data
		}
	}

	// Reader goroutine: detects disconnects and keeps the pong deadline fresh.
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn.SetReadLimit(512)
		_ = conn.SetReadDeadline(time.Now().Add(pongWait))
		conn.SetPongHandler(func(string) error {
			return conn.SetReadDeadline(time.Now().Add(pongWait))
		})
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	ticker := time.NewTicker(pingEvery)
	defer ticker.Stop()
	for {
		select {
		case msg := <-client.Send:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-done:
			return
		}
	}
}
