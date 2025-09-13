package websocket

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/models"
)

// MockServer implements the Server interface for testing
type MockServer struct {
	commands      []models.CommandMessage
	callbacks     []models.EventCallback
	serverInfo    models.ServerInfoMessage
	commandError  error
	commandResult interface{}
	mu            sync.Mutex
}

func NewMockServer() *MockServer {
	return &MockServer{
		commands:  make([]models.CommandMessage, 0),
		callbacks: make([]models.EventCallback, 0),
		serverInfo: models.ServerInfoMessage{
			FabricID:      1,
			SchemaVersion: 11,
			SDKVersion:    "test-1.0.0",
		},
		commandResult: map[string]interface{}{"status": "ok"},
	}
}

func (ms *MockServer) HandleCommand(ctx context.Context, cmd models.CommandMessage) (interface{}, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.commands = append(ms.commands, cmd)
	if ms.commandError != nil {
		return nil, ms.commandError
	}
	return ms.commandResult, nil
}

func (ms *MockServer) Subscribe(callback models.EventCallback) func() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.callbacks = append(ms.callbacks, callback)
	return func() {}
}

func (ms *MockServer) GetServerInfo() models.ServerInfoMessage {
	return ms.serverInfo
}

func (ms *MockServer) GetCommands() []models.CommandMessage {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return ms.commands
}

func (ms *MockServer) GetCallbackCount() int {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	return len(ms.callbacks)
}

func (ms *MockServer) SetCommandError(err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.commandError = err
}

func TestNewHandler(t *testing.T) {
	mockServer := NewMockServer()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	handler := NewHandler(mockServer, log)

	if handler.server != mockServer {
		t.Error("Server not set correctly")
	}
	if handler.logger != log {
		t.Error("Logger not set correctly")
	}
	if handler.connections == nil {
		t.Error("Connections map not initialized")
	}
	if len(handler.connections) != 0 {
		t.Error("Connections map should start empty")
	}
}

func TestBasicHandlerFunctionality(t *testing.T) {
	mockServer := NewMockServer()
	log := logger.NewConsoleLogger(logger.ErrorLevel)
	handler := NewHandler(mockServer, log)

	// Test connection count
	if handler.GetConnectionCount() != 0 {
		t.Errorf("Expected 0 connections initially, got %d", handler.GetConnectionCount())
	}

	// Test event broadcasting with no connections
	event := models.EventMessage{
		Event: models.EventTypeNodeAdded,
		Data:  map[string]interface{}{"test": "data"},
	}

	// Should not panic with no connections
	handler.BroadcastEvent(event)

	// Test shutdown with no connections
	handler.Shutdown()

	// Should still have 0 connections after shutdown
	if handler.GetConnectionCount() != 0 {
		t.Errorf("Expected 0 connections after shutdown, got %d", handler.GetConnectionCount())
	}
}

func TestMockServerFunctionality(t *testing.T) {
	mockServer := NewMockServer()

	// Test initial state
	if len(mockServer.GetCommands()) != 0 {
		t.Error("Expected no commands initially")
	}
	if mockServer.GetCallbackCount() != 0 {
		t.Error("Expected no callbacks initially")
	}

	// Test server info
	info := mockServer.GetServerInfo()
	if info.FabricID != 1 {
		t.Errorf("Expected FabricID 1, got %d", info.FabricID)
	}

	// Test command handling
	cmd := models.CommandMessage{
		MessageID: "test-123",
		Command:   "test_command",
	}

	result, err := mockServer.HandleCommand(context.Background(), cmd)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if result == nil {
		t.Error("Expected non-nil result")
	}

	// Test command was recorded
	commands := mockServer.GetCommands()
	if len(commands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(commands))
	}
	if commands[0].MessageID != cmd.MessageID {
		t.Errorf("Expected command ID %s, got %s", cmd.MessageID, commands[0].MessageID)
	}

	// Test error handling
	mockServer.SetCommandError(errors.New("test error"))
	_, err = mockServer.HandleCommand(context.Background(), cmd)
	if err == nil {
		t.Error("Expected error but got none")
	}

	// Test subscription
	callback := func(eventType models.EventType, data interface{}) {}

	unsubscribe := mockServer.Subscribe(callback)
	if mockServer.GetCallbackCount() != 1 {
		t.Error("Expected callback to be registered")
	}

	// Test unsubscribe function exists
	if unsubscribe == nil {
		t.Error("Expected unsubscribe function")
	}
	unsubscribe() // Should not panic
}
