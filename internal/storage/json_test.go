package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/models"
)

func TestNewJSONStorage(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)

	if storage.basePath != tempDir {
		t.Errorf("Expected basePath %s, got %s", tempDir, storage.basePath)
	}
	if storage.logger != log {
		t.Error("Logger not set correctly")
	}
	if storage.nodes == nil {
		t.Error("Nodes map not initialized")
	}
	if storage.vendors == nil {
		t.Error("Vendors map not initialized")
	}
	if storage.settings == nil {
		t.Error("Settings map not initialized")
	}
}

func TestJSONStorageStartStop(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)

	// Test start
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		t.Error("Storage directory was not created")
	}

	// Test stop
	err = storage.Stop()
	if err != nil {
		t.Fatalf("Failed to stop storage: %v", err)
	}
}

func TestNodeOperations(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer storage.Stop()

	// Test saving a node
	now := time.Now()
	node := &models.MatterNodeData{
		NodeID:           123,
		DateCommissioned: now,
		LastInterview:    now,
		InterviewVersion: 1,
		Available:        true,
		IsBridge:         false,
		Attributes:       map[string]interface{}{"test": "value"},
		AttributeSubscriptions: []models.AttributeSubscription{
			{EndpointID: intPtr(1), ClusterID: intPtr(2), AttributeID: intPtr(3)},
		},
	}

	err = storage.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	// Test getting the node
	retrievedNode, err := storage.GetNode(123)
	if err != nil {
		t.Fatalf("Failed to get node: %v", err)
	}

	if retrievedNode.NodeID != node.NodeID {
		t.Errorf("Expected NodeID %d, got %d", node.NodeID, retrievedNode.NodeID)
	}
	if retrievedNode.Available != node.Available {
		t.Errorf("Expected Available %v, got %v", node.Available, retrievedNode.Available)
	}

	// Test getting all nodes
	nodes, err := storage.GetNodes()
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}

	if len(nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(nodes))
	}

	// Test getting non-existent node
	_, err = storage.GetNode(999)
	if err == nil {
		t.Error("Expected error for non-existent node")
	}

	// Test deleting node
	err = storage.DeleteNode(123)
	if err != nil {
		t.Fatalf("Failed to delete node: %v", err)
	}

	// Verify node was deleted
	_, err = storage.GetNode(123)
	if err == nil {
		t.Error("Expected error after deleting node")
	}

	// Verify nodes list is empty
	nodes, err = storage.GetNodes()
	if err != nil {
		t.Fatalf("Failed to get nodes after deletion: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes after deletion, got %d", len(nodes))
	}
}

func TestVendorOperations(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer storage.Stop()

	// Test saving a vendor
	vendor := &models.VendorInfo{
		VendorID:             123,
		VendorName:           "Test Vendor",
		CompanyLegalName:     "Test Company Inc.",
		CompanyPreferredName: "Test Company",
		VendorLandingPageURL: "https://example.com",
		Creator:              "Test Creator",
	}

	err = storage.SaveVendor(vendor)
	if err != nil {
		t.Fatalf("Failed to save vendor: %v", err)
	}

	// Test getting the vendor
	retrievedVendor, err := storage.GetVendor(123)
	if err != nil {
		t.Fatalf("Failed to get vendor: %v", err)
	}

	if retrievedVendor.VendorID != vendor.VendorID {
		t.Errorf("Expected VendorID %d, got %d", vendor.VendorID, retrievedVendor.VendorID)
	}
	if retrievedVendor.VendorName != vendor.VendorName {
		t.Errorf("Expected VendorName %s, got %s", vendor.VendorName, retrievedVendor.VendorName)
	}

	// Test getting all vendors
	vendors, err := storage.GetVendors()
	if err != nil {
		t.Fatalf("Failed to get vendors: %v", err)
	}

	if len(vendors) != 1 {
		t.Errorf("Expected 1 vendor, got %d", len(vendors))
	}

	// Test getting non-existent vendor
	_, err = storage.GetVendor(999)
	if err == nil {
		t.Error("Expected error for non-existent vendor")
	}
}

func TestSettingsOperations(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer storage.Stop()

	// Test saving settings
	err = storage.SaveSetting("test_key", "test_value")
	if err != nil {
		t.Fatalf("Failed to save setting: %v", err)
	}

	err = storage.SaveSetting("number_key", 42)
	if err != nil {
		t.Fatalf("Failed to save number setting: %v", err)
	}

	// Test getting settings
	value, err := storage.GetSetting("test_key")
	if err != nil {
		t.Fatalf("Failed to get setting: %v", err)
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}

	numberValue, err := storage.GetSetting("number_key")
	if err != nil {
		t.Fatalf("Failed to get number setting: %v", err)
	}

	// JSON unmarshaling may convert numbers to float64, but our storage preserves int type
	// Check if it's an int or float64
	switch v := numberValue.(type) {
	case int:
		if v != 42 {
			t.Errorf("Expected int 42, got %v", v)
		}
	case float64:
		if v != float64(42) {
			t.Errorf("Expected float64(42), got %v", v)
		}
	default:
		t.Errorf("Expected int or float64, got %T: %v", numberValue, numberValue)
	}

	// Test getting non-existent setting
	_, err = storage.GetSetting("non_existent")
	if err == nil {
		t.Error("Expected error for non-existent setting")
	}

	// Test deleting setting
	err = storage.DeleteSetting("test_key")
	if err != nil {
		t.Fatalf("Failed to delete setting: %v", err)
	}

	// Verify setting was deleted
	_, err = storage.GetSetting("test_key")
	if err == nil {
		t.Error("Expected error after deleting setting")
	}
}

func TestStoragePersistence(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	// Create storage and add data
	storage1 := NewJSONStorage(tempDir, log)
	err := storage1.Start()
	if err != nil {
		t.Fatalf("Failed to start first storage: %v", err)
	}

	node := &models.MatterNodeData{
		NodeID:           456,
		DateCommissioned: time.Now(),
		LastInterview:    time.Now(),
		InterviewVersion: 1,
		Available:        true,
		IsBridge:         false,
		Attributes:       map[string]interface{}{"persistent": "test"},
	}

	err = storage1.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node in first storage: %v", err)
	}

	err = storage1.SaveSetting("persistent_setting", "persistent_value")
	if err != nil {
		t.Fatalf("Failed to save setting in first storage: %v", err)
	}

	err = storage1.Stop()
	if err != nil {
		t.Fatalf("Failed to stop first storage: %v", err)
	}

	// Create new storage instance and verify data persists
	storage2 := NewJSONStorage(tempDir, log)
	err = storage2.Start()
	if err != nil {
		t.Fatalf("Failed to start second storage: %v", err)
	}
	defer storage2.Stop()

	retrievedNode, err := storage2.GetNode(456)
	if err != nil {
		t.Fatalf("Failed to get persisted node: %v", err)
	}

	if retrievedNode.NodeID != node.NodeID {
		t.Errorf("Persisted node NodeID mismatch: expected %d, got %d", node.NodeID, retrievedNode.NodeID)
	}

	persistedValue, err := storage2.GetSetting("persistent_setting")
	if err != nil {
		t.Fatalf("Failed to get persisted setting: %v", err)
	}

	if persistedValue != "persistent_value" {
		t.Errorf("Persisted setting mismatch: expected 'persistent_value', got %v", persistedValue)
	}
}

func TestStorageSync(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer storage.Stop()

	// Add some data
	node := &models.MatterNodeData{
		NodeID:           789,
		DateCommissioned: time.Now(),
		LastInterview:    time.Now(),
		InterviewVersion: 1,
		Available:        true,
	}

	err = storage.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	// Test explicit sync
	err = storage.Sync()
	if err != nil {
		t.Fatalf("Failed to sync storage: %v", err)
	}

	// Verify files exist
	nodesFile := filepath.Join(tempDir, "nodes.json")
	if _, err := os.Stat(nodesFile); os.IsNotExist(err) {
		t.Error("Nodes file not created after sync")
	}
}

func TestBackupData(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage: %v", err)
	}
	defer storage.Stop()

	// Add some data
	node := &models.MatterNodeData{
		NodeID:           999,
		DateCommissioned: time.Now(),
		LastInterview:    time.Now(),
		InterviewVersion: 1,
		Available:        true,
	}

	err = storage.SaveNode(node)
	if err != nil {
		t.Fatalf("Failed to save node: %v", err)
	}

	// Create backup
	err = storage.BackupData()
	if err != nil {
		t.Fatalf("Failed to backup data: %v", err)
	}

	// Verify backup directory exists
	entries, err := os.ReadDir(tempDir)
	if err != nil {
		t.Fatalf("Failed to read temp directory: %v", err)
	}

	backupFound := false
	for _, entry := range entries {
		if entry.IsDir() && strings.Contains(entry.Name(), "backup_") {
			backupFound = true
			break
		}
	}

	if !backupFound {
		t.Error("Backup directory not created")
	}
}

func TestLoadNonExistentFiles(t *testing.T) {
	tempDir := t.TempDir()
	log := logger.NewConsoleLogger(logger.InfoLevel)

	storage := NewJSONStorage(tempDir, log)

	// Start should succeed even with no existing files
	err := storage.Start()
	if err != nil {
		t.Fatalf("Failed to start storage with no existing files: %v", err)
	}
	defer storage.Stop()

	// Should have empty collections
	nodes, err := storage.GetNodes()
	if err != nil {
		t.Fatalf("Failed to get nodes from empty storage: %v", err)
	}
	if len(nodes) != 0 {
		t.Errorf("Expected 0 nodes in empty storage, got %d", len(nodes))
	}

	vendors, err := storage.GetVendors()
	if err != nil {
		t.Fatalf("Failed to get vendors from empty storage: %v", err)
	}
	if len(vendors) != 0 {
		t.Errorf("Expected 0 vendors in empty storage, got %d", len(vendors))
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}
