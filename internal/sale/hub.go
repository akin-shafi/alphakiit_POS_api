package sale

import (
	"encoding/json"
	"log"
	"sync"
	
	"github.com/gofiber/websocket/v2"
)

// SaleEventType defines the types of events we broadcast
type SaleEventType string

const (
	EventOrderCreated SaleEventType = "ORDER_CREATED"
	EventOrderUpdated SaleEventType = "ORDER_UPDATED"
	EventOrderVoided  SaleEventType = "ORDER_VOIDED"
	EventOrderPaid    SaleEventType = "ORDER_PAID"
)

// SaleEvent represents the payload sent over WebSockets
type SaleEvent struct {
	Type       SaleEventType `json:"type"`
	BusinessID uint          `json:"business_id"`
	Data       interface{}   `json:"data"`
}

// KDSWebsocketHub manages all active KDS connections
type KDSWebsocketHub struct {
	// Map BusinessID -> map of connections
	Clients    map[uint]map[*websocket.Conn]bool
	Broadcast  chan SaleEvent
	Register   chan *KDSConn
	Unregister chan *KDSConn
	mu         sync.RWMutex
}

type KDSConn struct {
	BusinessID uint
	Conn       *websocket.Conn
}

var GlobalKDSHub *KDSWebsocketHub

func init() {
	GlobalKDSHub = &KDSWebsocketHub{
		Clients:    make(map[uint]map[*websocket.Conn]bool),
		Broadcast:  make(chan SaleEvent),
		Register:   make(chan *KDSConn),
		Unregister: make(chan *KDSConn),
	}
	go GlobalKDSHub.Run()
}

func (h *KDSWebsocketHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.BusinessID] == nil {
				h.Clients[client.BusinessID] = make(map[*websocket.Conn]bool)
			}
			h.Clients[client.BusinessID][client.Conn] = true
			h.mu.Unlock()
			log.Printf("[KDS] Client registered for business %d", client.BusinessID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if connections, ok := h.Clients[client.BusinessID]; ok {
				if _, ok := connections[client.Conn]; ok {
					delete(connections, client.Conn)
					client.Conn.Close()
					if len(connections) == 0 {
						delete(h.Clients, client.BusinessID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[KDS] Client unregistered for business %d", client.BusinessID)

		case event := <-h.Broadcast:
			h.mu.RLock()
			connections := h.Clients[event.BusinessID]
			if connections != nil {
				message, _ := json.Marshal(event)
				for conn := range connections {
					if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Printf("[KDS ERROR] Broadcast error: %v", err)
						conn.Close()
						// We don't unregister here to avoid deadlock, 
						// but usually you want a safer way to clean up
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastOrder sends an event to all connected KDS screens for a business
func (h *KDSWebsocketHub) BroadcastOrder(bizID uint, eventType SaleEventType, data interface{}) {
	h.Broadcast <- SaleEvent{
		Type:       eventType,
		BusinessID: bizID,
		Data:       data,
	}
}
