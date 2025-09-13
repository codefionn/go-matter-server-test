package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"

	"github.com/codefionn/go-matter-server/internal/bluetooth"
	"github.com/codefionn/go-matter-server/internal/config"
	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/mdns"
	"github.com/codefionn/go-matter-server/internal/models"
	"github.com/codefionn/go-matter-server/internal/storage"
	"github.com/codefionn/go-matter-server/internal/websocket"
)

// Server represents the main Matter server
type Server struct {
	config    *config.Config
	logger    *logger.Logger
	storage   storage.Storage
	wsHandler *websocket.Handler

	// Event system
	eventCallbacks []eventSubscription
	eventMu        sync.RWMutex

	// HTTP server
	httpServer *http.Server

	// mDNS server
	mdnsServer *mdns.Server
	mdnsZone   *mdns.MatterZone

	// Bluetooth manager (internal only)
	bluetoothManager *bluetooth.Manager

	// Matter-specific components
	nodes   map[int]*models.MatterNodeData
	nodesMu sync.RWMutex

	// Server info
	serverInfo models.ServerInfoMessage
}

// eventSubscription tracks a callback with an ID for safe unsubscribe
type eventSubscription struct {
	id string
	cb models.EventCallback
}

// New creates a new Matter server instance
func New(cfg *config.Config, log *logger.Logger) (*Server, error) {
	// Initialize storage
	jsonStorage := storage.NewJSONStorage(cfg.Storage.Path, log)

	s := &Server{
		config:  cfg,
		logger:  log,
		storage: jsonStorage,
		nodes:   make(map[int]*models.MatterNodeData),
		serverInfo: models.ServerInfoMessage{
			FabricID:                  cfg.Matter.FabricID,
			CompressedFabricID:        int64(cfg.Matter.FabricID), // Simplified for demo
			SchemaVersion:             11,                         // Current schema version
			MinSupportedSchemaVersion: 1,
			SDKVersion:                "go-matter-server-1.0.0",
			WiFiCredentialsSet:        false,
			ThreadCredentialsSet:      false,
			BluetoothEnabled:          false,
		},
	}

	// Initialize WebSocket handler
	s.wsHandler = websocket.NewHandler(s, log)

	// Initialize Bluetooth manager
	bluetoothLogger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	// Enable Bluetooth only when an adapter ID is provided (>= 0), mirroring python-matter-server.
	btEnabled := cfg.Bluetooth.AdapterID >= 0
	bluetoothConfig := bluetooth.Config{
		AdapterID:     fmt.Sprintf("hci%d", cfg.Bluetooth.AdapterID),
		Enabled:       btEnabled,
		EventCallback: s.EmitEvent,
		Logger:        bluetoothLogger,
	}

	if cfg.Bluetooth.AdapterID < 0 {
		bluetoothConfig.AdapterID = ""
	}

	var err error
	s.bluetoothManager, err = bluetooth.NewManager(bluetoothConfig)
	if err != nil {
		log.Warn("Failed to initialize Bluetooth manager", logger.ErrorField(err))
	}

	// Reflect actual Bluetooth availability in server info
	if s.bluetoothManager != nil && s.bluetoothManager.IsAvailable() {
		s.serverInfo.BluetoothEnabled = true
	} else {
		s.serverInfo.BluetoothEnabled = false
	}

	// Initialize mDNS if enabled
	if cfg.MDNS.Enabled {
		s.mdnsZone = mdns.NewMatterZone(cfg.MDNS.Hostname, log)

		// Try to determine primary interface
		var iface *net.Interface
		if cfg.Network.PrimaryInterface != "" {
			if i, err := net.InterfaceByName(cfg.Network.PrimaryInterface); err == nil {
				iface = i
			} else {
				log.Warn("Primary interface not found, using all interfaces",
					logger.String("interface", cfg.Network.PrimaryInterface),
					logger.ErrorField(err),
				)
			}
		}

		mdnsConfig := &mdns.Config{
			Interface: iface,
			Logger:    log,
			Zone:      s.mdnsZone,
		}

		var err error
		s.mdnsServer, err = mdns.NewServer(mdnsConfig)
		if err != nil {
			log.Warn("Failed to create mDNS server", logger.ErrorField(err))
		} else {
			log.Info("mDNS hostname advertisement enabled",
				logger.String("hostname", s.mdnsZone.GetHostname()),
			)
		}
	}

	return s, nil
}

// Run starts the server and blocks until shutdown
func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("Starting Matter server",
		logger.Int("port", s.config.Server.Port),
		logger.String("listen", strings.Join(s.config.Server.ListenAddresses, ", ")),
	)

	// Start storage
	if err := s.storage.Start(); err != nil {
		return fmt.Errorf("failed to start storage: %w", err)
	}
	defer s.storage.Stop()

	// Load existing nodes
	if err := s.loadNodes(); err != nil {
		s.logger.Error("Failed to load nodes", logger.ErrorField(err))
	}

	// Start mDNS server if enabled
	if s.mdnsServer != nil {
		if err := s.mdnsServer.Start(); err != nil {
			s.logger.Error("Failed to start mDNS server", logger.ErrorField(err))
		} else {
			s.logger.Info("mDNS server started",
				logger.String("hostname", s.mdnsZone.GetHostname()),
			)
		}
	}

	// Start Bluetooth manager if enabled
	if s.bluetoothManager != nil && s.bluetoothManager.IsEnabled() {
		if err := s.bluetoothManager.Start(); err != nil {
			s.logger.Error("Failed to start Bluetooth manager", logger.ErrorField(err))
		} else {
			s.logger.Info("Bluetooth manager started")
		}
	}

	// Setup HTTP router
	router := s.setupRouter()

	// Create HTTP server
	addr := fmt.Sprintf(":%d", s.config.Server.Port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	serverErr := make(chan error, 1)
	go func() {
		s.logger.Info("HTTP server listening", logger.String("addr", addr))
		if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
			serverErr <- err
		}
		close(serverErr)
	}()

	// Wait for context cancellation or server error
	select {
	case <-ctx.Done():
		s.logger.Info("Shutting down server...")
		return s.shutdown()
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	}
}

// HandleCommand processes WebSocket commands
func (s *Server) HandleCommand(ctx context.Context, cmd models.CommandMessage) (interface{}, error) {
	s.logger.Debug("Handling command",
		logger.String("command", cmd.Command),
		logger.String("message_id", cmd.MessageID),
	)

	switch models.APICommand(cmd.Command) {
	case models.APICommandServerInfo:
		return s.handleServerInfo()
	case models.APICommandGetNodes:
		return s.handleGetNodes()
	case models.APICommandGetNode:
		return s.handleGetNode(cmd.Args)
	case models.APICommandServerDiagnostics:
		return s.handleServerDiagnostics()
	case models.APICommandStartListening:
		return s.handleStartListening()
	case models.APICommandPingNode:
		return s.handlePingNode(cmd.Args)
	default:
		return nil, fmt.Errorf("unknown command: %s", cmd.Command)
	}
}

// Subscribe adds an event callback
func (s *Server) Subscribe(callback models.EventCallback) func() {
	s.eventMu.Lock()
	defer s.eventMu.Unlock()

	id := models.GenerateMessageID()
	s.eventCallbacks = append(s.eventCallbacks, eventSubscription{id: id, cb: callback})

	// Return unsubscribe function (removes by ID)
	return func() {
		s.eventMu.Lock()
		defer s.eventMu.Unlock()

		for i := range s.eventCallbacks {
			if s.eventCallbacks[i].id == id {
				s.eventCallbacks = append(s.eventCallbacks[:i], s.eventCallbacks[i+1:]...)
				break
			}
		}
	}
}

// GetServerInfo returns server information
func (s *Server) GetServerInfo() models.ServerInfoMessage {
	return s.serverInfo
}

// EmitEvent sends an event to all subscribers
func (s *Server) EmitEvent(eventType models.EventType, data interface{}) {
	s.eventMu.RLock()
	callbacks := make([]eventSubscription, len(s.eventCallbacks))
	copy(callbacks, s.eventCallbacks)
	s.eventMu.RUnlock()

	for _, sub := range callbacks {
		// Run callbacks asynchronously to avoid blocking
		go sub.cb(eventType, data)
	}
}

// Command handlers

func (s *Server) handleServerInfo() (interface{}, error) {
	return s.serverInfo, nil
}

func (s *Server) handleGetNodes() (interface{}, error) {
	s.nodesMu.RLock()
	defer s.nodesMu.RUnlock()

	nodes := make([]*models.MatterNodeData, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes, nil
}

func (s *Server) handleGetNode(args map[string]interface{}) (interface{}, error) {
	nodeID, err := parseNodeID(args)
	if err != nil {
		return nil, err
	}

	s.nodesMu.RLock()
	node, exists := s.nodes[nodeID]
	s.nodesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node %d not found", nodeID)
	}

	nodeCopy := *node
	return &nodeCopy, nil
}

func (s *Server) handleServerDiagnostics() (interface{}, error) {
	nodes, err := s.handleGetNodes()
	if err != nil {
		return nil, err
	}

	nodeList := nodes.([]*models.MatterNodeData)
	nodeSlice := make([]models.MatterNodeData, len(nodeList))
	for i, node := range nodeList {
		nodeSlice[i] = *node
	}

	return models.ServerDiagnostics{
		Info:   s.serverInfo,
		Nodes:  nodeSlice,
		Events: []interface{}{}, // Empty for now
	}, nil
}

func (s *Server) handleStartListening() (interface{}, error) {
	// Return all nodes for initial state
	return s.handleGetNodes()
}

func (s *Server) handlePingNode(args map[string]interface{}) (interface{}, error) {
	nodeID, err := parseNodeID(args)
	if err != nil {
		return nil, err
	}

	// Simple ping implementation - in real implementation this would ping the actual device
	s.nodesMu.RLock()
	_, exists := s.nodes[nodeID]
	s.nodesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("node %d not found", nodeID)
	}

	result := models.NodePingResult{
		"reachable": true,
	}

	return result, nil
}

// Bluetooth command handlers removed: the Go server does not expose
// standalone Bluetooth management endpoints to mirror python-matter-server.

// HTTP handlers

func (s *Server) setupRouter() *mux.Router {
	router := mux.NewRouter()

	// WebSocket endpoint
	router.HandleFunc("/ws", s.wsHandler.HandleWebSocket)

	// HTTP API endpoints
	api := router.PathPrefix("/api").Subrouter()
	api.HandleFunc("/info", s.handleInfoHTTP).Methods("GET")
	api.HandleFunc("/nodes", s.handleNodesHTTP).Methods("GET")
	api.HandleFunc("/diagnostics", s.handleDiagnosticsHTTP).Methods("GET")

	// Health check
	router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Serve static files if available - removed for now as ServeStatic field doesn't exist
	// TODO: Add static file serving configuration

	// Add middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)

	return router
}

func (s *Server) handleInfoHTTP(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, s.serverInfo)
}

func (s *Server) handleNodesHTTP(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.handleGetNodes()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, nodes)
}

func (s *Server) handleDiagnosticsHTTP(w http.ResponseWriter, r *http.Request) {
	diagnostics, err := s.handleServerDiagnostics()
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.writeJSON(w, diagnostics)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":      "ok",
		"timestamp":   time.Now().UTC(),
		"connections": s.wsHandler.GetConnectionCount(),
		"nodes":       len(s.nodes),
	}

	s.writeJSON(w, health)
}

// Helper methods

func (s *Server) loadNodes() error {
	nodes, err := s.storage.GetNodes()
	if err != nil {
		return fmt.Errorf("failed to load nodes from storage: %w", err)
	}

	s.nodesMu.Lock()
	defer s.nodesMu.Unlock()

	for _, node := range nodes {
		s.nodes[node.NodeID] = node
	}

	s.logger.Info("Loaded nodes from storage", logger.Int("count", len(nodes)))
	return nil
}

func (s *Server) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown HTTP server", logger.ErrorField(err))
	}

	// Shutdown WebSocket handler
	s.wsHandler.Shutdown()

	// Shutdown Bluetooth manager
	if s.bluetoothManager != nil {
		if err := s.bluetoothManager.Stop(); err != nil {
			s.logger.Error("Failed to shutdown Bluetooth manager", logger.ErrorField(err))
		}
	}

	// Shutdown mDNS server
	if s.mdnsServer != nil {
		if err := s.mdnsServer.Shutdown(); err != nil {
			s.logger.Error("Failed to shutdown mDNS server", logger.ErrorField(err))
		}
	}

	// Emit shutdown event
	s.EmitEvent(models.EventTypeServerShutdown, nil)

	s.logger.Info("Server shutdown complete")
	return nil
}

// Middleware and utilities will go in separate files
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		s.logger.Info("HTTP request",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.Duration("duration", duration),
			logger.String("remote_addr", r.RemoteAddr),
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", logger.ErrorField(err))
	}
}

func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errorResponse := map[string]interface{}{
		"error":     message,
		"code":      code,
		"timestamp": time.Now().UTC(),
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		s.logger.Error("Failed to encode error response", logger.ErrorField(err))
	} else {
		s.logger.Warn("HTTP error response",
			logger.Int("status", code),
			logger.String("message", message),
		)
	}
}

// parseNodeID extracts a node_id from args, accepting number or string
func parseNodeID(args map[string]interface{}) (int, error) {
	v, ok := args["node_id"]
	if !ok {
		return 0, fmt.Errorf("missing required parameter: node_id")
	}
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case int:
		return t, nil
	case int64:
		return int(t), nil
	case json.Number:
		// If caller used Decoder.UseNumber
		i, err := t.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid node_id: %v", err)
		}
		return int(i), nil
	case string:
		// Try to parse string number
		var n json.Number = json.Number(t)
		i, err := n.Int64()
		if err != nil {
			return 0, fmt.Errorf("invalid node_id: expected number, got %q", t)
		}
		return int(i), nil
	default:
		return 0, fmt.Errorf("invalid node_id: unsupported type")
	}
}
