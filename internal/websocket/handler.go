package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/models"
)

const (
	// WebSocket configuration
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024 // 1MB
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Handler manages WebSocket connections and message routing
type Handler struct {
	server        Server
	logger        *logger.Logger
	connections   map[string]*Connection
	connectionsMu sync.RWMutex
}

// Server interface defines the methods the WebSocket handler needs
type Server interface {
	HandleCommand(ctx context.Context, cmd models.CommandMessage) (interface{}, error)
	Subscribe(callback models.EventCallback) func()
	GetServerInfo() models.ServerInfoMessage
}

// Connection represents a WebSocket client connection
type Connection struct {
	id          string
	conn        *websocket.Conn
	handler     *Handler
	send        chan []byte
	ctx         context.Context
	cancel      context.CancelFunc
	logger      *logger.Logger
	unsubscribe func()
}

// NewHandler creates a new WebSocket handler
func NewHandler(server Server, log *logger.Logger) *Handler {
	return &Handler{
		server:      server,
		logger:      log,
		connections: make(map[string]*Connection),
	}
}

// HandleWebSocket handles WebSocket upgrade requests
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", logger.ErrorField(err))
		return
	}

	connID := models.GenerateMessageID()
	ctx, cancel := context.WithCancel(r.Context())

	client := &Connection{
		id:      connID,
		conn:    conn,
		handler: h,
		send:    make(chan []byte, 256),
		ctx:     ctx,
		cancel:  cancel,
		logger:  h.logger.With(logger.String("connection", connID)),
	}

	// Subscribe to server events
	client.unsubscribe = h.server.Subscribe(client.handleEvent)

	// Register connection
	h.connectionsMu.Lock()
	h.connections[connID] = client
	h.connectionsMu.Unlock()

	client.logger.Info("WebSocket connection established")

	// Send server info immediately via direct WebSocket write
	serverInfo := h.server.GetServerInfo()
	data, err := json.Marshal(serverInfo)
	if err != nil {
		client.logger.Error("Failed to marshal server info", logger.ErrorField(err))
		client.close()
		return
	}

	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		client.logger.Error("Failed to send server info", logger.ErrorField(err))
		client.close()
		return
	}

	// Start goroutines for this connection
	go client.writePump()
	go client.readPump()
}

// BroadcastEvent sends an event to all connected clients
func (h *Handler) BroadcastEvent(event models.EventMessage) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("Failed to marshal event", logger.ErrorField(err))
		return
	}

	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()

	for _, conn := range h.connections {
		select {
		case conn.send <- data:
		default:
			// Connection is busy, close it
			conn.close()
		}
	}
}

// GetConnectionCount returns the number of active connections
func (h *Handler) GetConnectionCount() int {
	h.connectionsMu.RLock()
	defer h.connectionsMu.RUnlock()
	return len(h.connections)
}

// Shutdown closes all connections
func (h *Handler) Shutdown() {
	h.connectionsMu.Lock()
	defer h.connectionsMu.Unlock()

	for _, conn := range h.connections {
		conn.close()
	}

	h.connections = make(map[string]*Connection)
	h.logger.Info("WebSocket handler shutdown")
}

// Connection methods

func (c *Connection) readPump() {
	defer func() {
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error", logger.ErrorField(err))
			}
			return
		}

		var cmd models.CommandMessage
		if err := json.Unmarshal(message, &cmd); err != nil {
			c.logger.Error("Failed to unmarshal command", logger.ErrorField(err))
			c.sendError(models.GenerateMessageID(), 400, "Invalid message format")
			continue
		}

		go c.handleCommand(cmd)
	}
}

func (c *Connection) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Connection) handleCommand(cmd models.CommandMessage) {
	c.logger.Debug("Handling command",
		logger.String("command", cmd.Command),
		logger.String("message_id", cmd.MessageID),
	)

	result, err := c.handler.server.HandleCommand(c.ctx, cmd)
	if err != nil {
		c.logger.Error("Command failed",
			logger.String("command", cmd.Command),
			logger.ErrorField(err),
		)
		c.sendError(cmd.MessageID, 500, err.Error())
		return
	}

	response := models.SuccessResultMessage{
		ResultMessageBase: models.ResultMessageBase{
			MessageID: cmd.MessageID,
		},
		Result: result,
	}

	if err := c.sendMessage(response); err != nil {
		c.logger.Error("Failed to send command response", logger.ErrorField(err))
	}
}

func (c *Connection) handleEvent(eventType models.EventType, data interface{}) {
	event := models.EventMessage{
		Event: eventType,
		Data:  data,
	}

	if err := c.sendMessage(event); err != nil {
		c.logger.Error("Failed to send event",
			logger.String("event", string(eventType)),
			logger.ErrorField(err),
		)
	}
}

func (c *Connection) sendMessage(msg interface{}) error {
	// Check if connection is already closed
	select {
	case <-c.ctx.Done():
		return fmt.Errorf("connection closed")
	default:
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Try to send with timeout
	select {
	case c.send <- data:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("connection closed")
	case <-time.After(1 * time.Second):
		return fmt.Errorf("send timeout")
	}
}

func (c *Connection) sendError(messageID string, code int, details string) {
	errorMsg := models.ErrorResultMessage{
		ResultMessageBase: models.ResultMessageBase{
			MessageID: messageID,
		},
		ErrorCode: code,
		Details:   &details,
	}

	if err := c.sendMessage(errorMsg); err != nil {
		c.logger.Error("Failed to send error message", logger.ErrorField(err))
	}
}

func (c *Connection) close() {
	// Only close once
	select {
	case <-c.ctx.Done():
		return // Already closed
	default:
	}

	if c.unsubscribe != nil {
		c.unsubscribe()
		c.unsubscribe = nil
	}

	c.cancel()

	// Remove from handler's connection map
	c.handler.connectionsMu.Lock()
	delete(c.handler.connections, c.id)
	c.handler.connectionsMu.Unlock()

	// Close send channel safely
	go func() {
		defer func() {
			if r := recover(); r != nil {
				// Channel already closed, ignore
			}
		}()
		close(c.send)
	}()

	c.conn.Close()

	c.logger.Info("WebSocket connection closed")
}
