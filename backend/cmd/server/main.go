package main


import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"bus-tracker/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var ctx = context.Background()

type TelemetryPayload struct {
	BusID     int     `json:"bus_id"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Speed     float64 `json:"speed"`
	Heading   float64 `json:"heading"`
}

type ClientManager struct {
	clients map[int]map[*websocket.Conn]bool
	mu      sync.Mutex
}

var manager = ClientManager{
	clients: make(map[int]map[*websocket.Conn]bool),
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	// Initialize PostgreSQL / PostGIS
	pgDB := db.ConnectPostgres()
	defer pgDB.Close()

	// Initialize Redis Client
	// Initialize Redis Client
	rdb := redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379",
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis successfully!")

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "running", "service": "bus-tracker-backend"})
	})

	// Ingest Telemetry and persist into PostGIS + Redis broadcast
	r.POST("/api/v1/telemetry", func(c *gin.Context) {
		var payload TelemetryPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Save telemetry log into PostGIS using spatial ST_SetSRID/ST_MakePoint
		query := `INSERT INTO live_telemetry_logs (bus_id, location, speed, heading) VALUES ($1, ST_SetSRID(ST_MakePoint($2, $3), 4326), $4, $5)`
		_, err := pgDB.Exec(query, payload.BusID, payload.Longitude, payload.Latitude, payload.Speed, payload.Heading)
		if err != nil {
			log.Printf("Failed to insert telemetry into database: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database write failed"})
			return
		}

		// Serialize payload and broadcast via Redis
		data, err := json.Marshal(payload)
		if err == nil {
			rdb.Publish(ctx, "bus_updates", data)
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Telemetry logged and broadcasted"})
	})

	log.Println("Backend server starting on port 8080...")
	r.Run(":8080")
}