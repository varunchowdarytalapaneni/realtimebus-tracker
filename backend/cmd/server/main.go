package main


import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
    "fmt"
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
	// Background worker to read from Redis Pub/Sub and broadcast to WebSockets
    go func() {
        pubsub := rdb.Subscribe(ctx, "bus_updates")
        defer pubsub.Close()
        ch := pubsub.Channel()
        for msg := range ch {
            var payload TelemetryPayload
            if err := json.Unmarshal([]byte(msg.Payload), &payload); err == nil {
                manager.mu.Lock()
                if busClients, exists := manager.clients[payload.BusID]; exists {
                    for client := range busClients {
                        if err := client.WriteMessage(websocket.TextMessage, []byte(msg.Payload)); err != nil {
                            client.Close()
                            delete(busClients, client)
                        }
                    }
                }
                manager.mu.Unlock()
            }
        }
    }()

	// WebSocket endpoint for real-time tracking dashboard / wscat
    r.GET("/ws/track/:id", func(c *gin.Context) {
        idStr := c.Param("id")
        var busID int
        if _, err := fmt.Sscanf(idStr, "%d", &busID); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bus ID"})
            return
        }

        conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
        if err != nil {
            log.Printf("Failed to upgrade websocket: %v", err)
            return
        }

        manager.mu.Lock()
        if manager.clients[busID] == nil {
            manager.clients[busID] = make(map[*websocket.Conn]bool)
        }
        manager.clients[busID][conn] = true
        manager.mu.Unlock()

        defer func() {
            manager.mu.Lock()
            delete(manager.clients[busID], conn)
            if len(manager.clients[busID]) == 0 {
                delete(manager.clients, busID)
            }
            manager.mu.Unlock()
            conn.Close()
        }()

        // Keep connection open and read messages (handles client disconnects)
        for {
            _, _, err := conn.ReadMessage()
            if err != nil {
                break
            }
        }
    })
	r.Run(":8080")
}