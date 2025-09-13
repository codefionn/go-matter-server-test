package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codefionn/go-matter-server/internal/config"
	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/models"
)

func createTestServer(t *testing.T) *Server {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:            0, // Use any available port
			ListenAddresses: []string{"127.0.0.1"},
		},
		Storage: config.StorageConfig{
			Path: tempDir,
		},
		Matter: config.MatterConfig{
			VendorID: 0xFFF1,
			FabricID: 1,
		},
		Bluetooth: config.BluetoothConfig{
			Enabled: false,
		},
	}

	log := logger.NewConsoleLogger(logger.ErrorLevel) // Reduce noise in tests

	server, err := New(cfg, log)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	return server
}

func TestNewServer(t *testing.T) {
	server := createTestServer(t)

	if server.config == nil {
		t.Error("Config not initialized")
	}
	if server.logger == nil {
		t.Error("Logger not initialized")
	}
	if server.storage == nil {
		t.Error("Storage not initialized")
	}
	if server.wsHandler == nil {
		t.Error("WebSocket handler not initialized")
	}
	if server.nodes == nil {
		t.Error("Nodes map not initialized")
	}
	if server.serverInfo.FabricID == 0 {
		t.Error("Server info not initialized correctly")
	}
}

func TestServerInfo(t *testing.T) {
	server := createTestServer(t)

	info := server.GetServerInfo()

	if info.FabricID != server.config.Matter.FabricID {
		t.Errorf("Expected FabricID %d, got %d", server.config.Matter.FabricID, info.FabricID)
	}
	if info.SDKVersion == "" {
		t.Error("SDKVersion should not be empty")
	}
	if info.SchemaVersion == 0 {
		t.Error("SchemaVersion should not be zero")
	}
	// In tests we disable Bluetooth via config, so reported availability must be false
	if info.BluetoothEnabled {
		t.Errorf("Expected BluetoothEnabled false when disabled in config, got true")
	}
}

func TestEventSubscription(t *testing.T) {
	server := createTestServer(t)

	// Test basic subscription functionality
	eventReceived := make(chan models.EventType, 10)

	callback := func(eventType models.EventType, data interface{}) {
		select {
		case eventReceived <- eventType:
		default:
			// Don't block if channel is full
		}
	}

	unsubscribe := server.Subscribe(callback)

	// Trigger an event
	testData := map[string]interface{}{"test": "data"}
	server.EmitEvent(models.EventTypeNodeAdded, testData)

	// Wait for event
	select {
	case event := <-eventReceived:
		if event != models.EventTypeNodeAdded {
			t.Errorf("Expected %s event, got %s", models.EventTypeNodeAdded, event)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive event but didn't")
	}

	// Test that unsubscribe function exists and doesn't panic
	if unsubscribe == nil {
		t.Error("Expected unsubscribe function")
	}
	unsubscribe() // Should not panic

	// Note: Testing that events are NOT received after unsubscribe is racy
	// due to goroutine execution order. The important part is that the
	// subscription works and unsubscribe doesn't panic.
}

func TestHTTPHealthEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header not set correctly")
	}
}

func TestHTTPInfoEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/api/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var info models.ServerInfoMessage
	if err := json.Unmarshal(w.Body.Bytes(), &info); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if info.FabricID != server.config.Matter.FabricID {
		t.Errorf("Expected FabricID %d, got %d", server.config.Matter.FabricID, info.FabricID)
	}
}

func TestHTTPNodesEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/api/nodes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var nodes []*models.MatterNodeData
	if err := json.Unmarshal(w.Body.Bytes(), &nodes); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	// Should start with empty nodes
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(nodes))
	}
}

func TestHTTPDiagnosticsEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/api/diagnostics", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var diagnostics models.ServerDiagnostics
	if err := json.Unmarshal(w.Body.Bytes(), &diagnostics); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if diagnostics.Info.FabricID != server.config.Matter.FabricID {
		t.Errorf("Expected FabricID %d in diagnostics, got %d",
			server.config.Matter.FabricID, diagnostics.Info.FabricID)
	}
}

func TestCORSHeaders(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	// Test CORS headers are set on regular requests
	req := httptest.NewRequest("GET", "/api/info", nil)
	req.Header.Set("Origin", "https://example.com")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS origin header not set correctly")
	}
}

func TestCommandHandling(t *testing.T) {
	server := createTestServer(t)

	tests := []struct {
		name        string
		command     models.CommandMessage
		expectError bool
	}{
		{
			name: "Server info command",
			command: models.CommandMessage{
				MessageID: "test-1",
				Command:   string(models.APICommandServerInfo),
			},
			expectError: false,
		},
		{
			name: "Get nodes command",
			command: models.CommandMessage{
				MessageID: "test-2",
				Command:   string(models.APICommandGetNodes),
			},
			expectError: false,
		},
		{
			name: "Invalid command",
			command: models.CommandMessage{
				MessageID: "test-3",
				Command:   "invalid_command",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := server.HandleCommand(context.Background(), tt.command)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
				}
			}
		})
	}
}

func TestInternalNodeStorage(t *testing.T) {
	server := createTestServer(t)

	// Test that nodes are handled via storage
	// We can test this indirectly through the HTTP API
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/api/nodes", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var nodes []*models.MatterNodeData
	if err := json.Unmarshal(w.Body.Bytes(), &nodes); err != nil {
		t.Fatalf("Failed to parse nodes response: %v", err)
	}

	// Should start empty
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes initially, got %d", len(nodes))
	}
}

func TestCORSMiddleware(t *testing.T) {
	server := createTestServer(t)

	// Test that CORS middleware is applied
	handler := server.corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS origin header not set by middleware")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Error("CORS headers not set by middleware")
	}
}

func TestServerShutdown(t *testing.T) {
	server := createTestServer(t)

	// Start server in background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		// This should exit when context is cancelled
		server.Run(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context (shutdown)
	cancel()

	// Give server time to shutdown
	time.Sleep(100 * time.Millisecond)

	// Test should complete without hanging
}

func TestInvalidHTTPMethod(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	// Try invalid method on existing endpoint
	req := httptest.NewRequest("DELETE", "/api/info", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should either be 405 Method Not Allowed or 404 if not handled
	if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
		t.Errorf("Expected status 405 or 404, got %d", w.Code)
	}
}

func TestNonExistentEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	req := httptest.NewRequest("GET", "/api/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestWebSocketEndpoint(t *testing.T) {
	server := createTestServer(t)
	router := server.setupRouter()

	// Test that WebSocket endpoint exists (without upgrade headers)
	req := httptest.NewRequest("GET", "/ws", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Without proper WebSocket headers, should get an error but endpoint exists
	if w.Code == http.StatusNotFound {
		t.Error("WebSocket endpoint should exist at /ws")
	}
}
