package printing

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type PrintJob struct {
	PrinterID uint        `json:"printer_id"`
	Content   string      `json:"content"` // Raw ESC/POS or specific format
	Data      interface{} `json:"data"`    // Structured data if the agent handles layout
}

type PrintEvent struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type PrintingHub struct {
	// Map OutletID -> map of connections (usually one agent per outlet, but support multiple)
	Clients    map[uint]map[*websocket.Conn]bool
	Register   chan *AgentConn
	Unregister chan *AgentConn
	mu         sync.RWMutex
}

type AgentConn struct {
	OutletID uint
	Conn     *websocket.Conn
}

var GlobalPrintingHub *PrintingHub

func init() {
	GlobalPrintingHub = &PrintingHub{
		Clients:    make(map[uint]map[*websocket.Conn]bool),
		Register:   make(chan *AgentConn),
		Unregister: make(chan *AgentConn),
	}
	go GlobalPrintingHub.Run()
}

func (h *PrintingHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.Clients[client.OutletID] == nil {
				h.Clients[client.OutletID] = make(map[*websocket.Conn]bool)
			}
			h.Clients[client.OutletID][client.Conn] = true
			h.mu.Unlock()
			log.Printf("[PRINT] Agent registered for outlet %d", client.OutletID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if connections, ok := h.Clients[client.OutletID]; ok {
				if _, ok := connections[client.Conn]; ok {
					delete(connections, client.Conn)
					client.Conn.Close()
					if len(connections) == 0 {
						delete(h.Clients, client.OutletID)
					}
				}
			}
			h.mu.Unlock()
			log.Printf("[PRINT] Agent unregistered for outlet %d", client.OutletID)
		}
	}
}

// SendJobToOutlet sends a print job to all agents connected for a specific outlet
func (h *PrintingHub) SendJobToOutlet(outletID uint, job PrintJob) {
	h.mu.RLock()
	connections := h.Clients[outletID]
	if connections != nil {
		message, _ := json.Marshal(PrintEvent{
			Type: "PRINT_JOB",
			Data: job,
		})
		for conn := range connections {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("[PRINT ERROR] Sending job error: %v", err)
				// We don't unregister here to avoid deadlock
			}
		}
	}
	h.mu.RUnlock()
}
