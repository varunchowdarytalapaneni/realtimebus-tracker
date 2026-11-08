package hub

import (
	"context"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"

	"bus-tracker/internal/db"
)

type Client struct {
	BusID int64
	Send  chan []byte
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[int64]map[*Client]struct{}
}

func New() *Hub {
	return &Hub{rooms: make(map[int64]map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.BusID] == nil {
		h.rooms[c.BusID] = make(map[*Client]struct{})
	}
	h.rooms[c.BusID][c] = struct{}{}
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if room := h.rooms[c.BusID]; room != nil {
		delete(room, c)
		if len(room) == 0 {
			delete(h.rooms, c.BusID)
		}
	}
}

func (h *Hub) Broadcast(busID int64, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.rooms[busID] {
		select {
		case c.Send <- msg:
		default:
		}
	}
}

func (h *Hub) Run(ctx context.Context, in <-chan *redis.Message) {
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-in:
			if !ok {
				return
			}
			id, err := strconv.ParseInt(strings.TrimPrefix(m.Channel, db.ChannelPrefix), 10, 64)
			if err == nil {
				h.Broadcast(id, []byte(m.Payload))
			}
		}
	}
}
