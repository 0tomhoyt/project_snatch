package notification

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"github.com/yourname/harborlink/internal/model"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections for development
		// In production, implement proper origin checking
		return true
	},
}

// Client represents a WebSocket client connection
type Client struct {
	hub      *WebSocketHub
	conn     *websocket.Conn
	send     chan []byte
	tenantID string
	mu       sync.Mutex
}

// WebSocketHub maintains active WebSocket connections and broadcasts messages
type WebSocketHub struct {
	clients    map[string]map[*Client]bool // tenantID -> clients
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	mu         sync.RWMutex
}

// BroadcastMessage represents a message to broadcast
type BroadcastMessage struct {
	TenantID string
	Message  interface{}
}

// NewWebSocketHub creates a new WebSocket hub
func NewWebSocketHub() *WebSocketHub {
	return &WebSocketHub{
		clients:    make(map[string]map[*Client]bool),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
		broadcast:  make(chan *BroadcastMessage, 1024),
	}
}

// Run starts the hub's main loop
func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient registers a new client
func (h *WebSocketHub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.tenantID] == nil {
		h.clients[client.tenantID] = make(map[*Client]bool)
	}
	h.clients[client.tenantID][client] = true

	log.Printf("[INFO] WebSocket: client connected for tenant %s", client.tenantID)
}

// unregisterClient unregisters a client
func (h *WebSocketHub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.clients[client.tenantID]; ok {
		if _, ok := clients[client]; ok {
			delete(clients, client)
			close(client.send)
			if len(clients) == 0 {
				delete(h.clients, client.tenantID)
			}
		}
	}

	log.Printf("[INFO] WebSocket: client disconnected for tenant %s", client.tenantID)
}

// broadcastMessage broadcasts a message to all clients of a tenant
func (h *WebSocketHub) broadcastMessage(message *BroadcastMessage) {
	h.mu.RLock()
	clients, ok := h.clients[message.TenantID]
	h.mu.RUnlock()

	if !ok {
		return
	}

	data, err := json.Marshal(message.Message)
	if err != nil {
		log.Printf("[ERROR] WebSocket: failed to marshal message: %v", err)
		return
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// Client buffer full, close connection
			close(client.send)
			h.mu.Lock()
			delete(h.clients[message.TenantID], client)
			h.mu.Unlock()
		}
	}
}

// SendToTenant sends a message to all clients of a specific tenant
func (h *WebSocketHub) SendToTenant(tenantID string, message interface{}) error {
	h.broadcast <- &BroadcastMessage{
		TenantID: tenantID,
		Message:  message,
	}
	return nil
}

// GetClientCount returns the number of connected clients for a tenant
func (h *WebSocketHub) GetClientCount(tenantID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[tenantID])
}

// GetTotalClientCount returns the total number of connected clients
func (h *WebSocketHub) GetTotalClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	total := 0
	for _, clients := range h.clients {
		total += len(clients)
	}
	return total
}

// NewClient creates a new WebSocket client
func NewClient(hub *WebSocketHub, conn *websocket.Conn, tenantID string) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		tenantID: tenantID,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[ERROR] WebSocket: read error: %v", err)
			}
			break
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWebSocket handles WebSocket requests
func ServeWebSocket(hub *WebSocketHub, tenantID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("[ERROR] WebSocket: upgrade failed: %v", err)
			return
		}

		client := NewClient(hub, conn, tenantID)
		hub.register <- client

		// Start read/write pumps
		go client.WritePump()
		go client.ReadPump()
	}
}

// NotifierService implements the Notifier interface using WebSocket
type NotifierService struct {
	hub *WebSocketHub
}

// NewNotifierService creates a new notifier service
func NewNotifierService(hub *WebSocketHub) *NotifierService {
	return &NotifierService{hub: hub}
}

// NotifySlotOpened notifies a client that a slot has opened
func (n *NotifierService) NotifySlotOpened(ctx context.Context, tenantID string, watch *model.SlotWatch, slot *model.SlotStatus) error {
	message := model.WebSocketMessage{
		Type:           "SLOT_OPENED",
		WatchReference: watch.Reference,
		Carrier:        slot.CarrierCode,
		Slot: &model.SlotInfo{
			CarrierCode:   slot.CarrierCode,
			VesselName:    slot.VesselName,
			VoyageNumber:  slot.VoyageNumber,
			POL:           slot.POL,
			POD:           slot.POD,
			ETD:           slot.ETD,
			ETA:           slot.ETA,
			EquipmentType: slot.EquipmentType,
			Available:     slot.Available,
			AvailableQty:  slot.AvailableQty,
		},
		Timestamp: time.Now(),
	}

	return n.hub.SendToTenant(tenantID, message)
}

// NotifyLockResult notifies a client of a lock result
func (n *NotifierService) NotifyLockResult(ctx context.Context, tenantID string, watch *model.SlotWatch, success bool, bookingRef string, errMsg string) error {
	msgType := "LOCK_SUCCESS"
	if !success {
		msgType = "LOCK_FAILED"
	}

	message := model.WebSocketMessage{
		Type:           msgType,
		WatchReference: watch.Reference,
		Carrier:        watch.TriggeredByCarrier,
		BookingRef:     bookingRef,
		Error:          errMsg,
		Timestamp:      time.Now(),
	}

	return n.hub.SendToTenant(tenantID, message)
}

// NotifyWatchCancelled notifies a client that a watch was cancelled
func (n *NotifierService) NotifyWatchCancelled(ctx context.Context, tenantID string, watch *model.SlotWatch) error {
	message := model.WebSocketMessage{
		Type:           "WATCH_CANCELLED",
		WatchReference: watch.Reference,
		Timestamp:      time.Now(),
	}

	return n.hub.SendToTenant(tenantID, message)
}

// NotifyWatchExpired notifies a client that a watch has expired
func (n *NotifierService) NotifyWatchExpired(ctx context.Context, tenantID string, watch *model.SlotWatch) error {
	message := model.WebSocketMessage{
		Type:           "WATCH_EXPIRED",
		WatchReference: watch.Reference,
		Timestamp:      time.Now(),
	}

	return n.hub.SendToTenant(tenantID, message)
}

// NotifySlotClosed notifies a client that a slot has closed
func (n *NotifierService) NotifySlotClosed(ctx context.Context, tenantID string, watch *model.SlotWatch, slot *model.SlotStatus) error {
	message := model.WebSocketMessage{
		Type:           "SLOT_CLOSED",
		WatchReference: watch.Reference,
		Carrier:        slot.CarrierCode,
		Slot: &model.SlotInfo{
			CarrierCode:   slot.CarrierCode,
			VesselName:    slot.VesselName,
			VoyageNumber:  slot.VoyageNumber,
			POL:           slot.POL,
			POD:           slot.POD,
			ETD:           slot.ETD,
		},
		Timestamp: time.Now(),
	}

	return n.hub.SendToTenant(tenantID, message)
}
