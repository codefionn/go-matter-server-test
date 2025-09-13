package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/models"
)

// JSONStorage implements storage using JSON files
type JSONStorage struct {
	basePath string
	logger   *logger.Logger
	mu       sync.RWMutex

	// In-memory cache
	nodes    map[int]*models.MatterNodeData
	vendors  map[int]*models.VendorInfo
	settings map[string]interface{}
}

// Storage interface defines storage operations
type Storage interface {
	// Node operations
	GetNode(nodeID int) (*models.MatterNodeData, error)
	GetNodes() ([]*models.MatterNodeData, error)
	SaveNode(node *models.MatterNodeData) error
	DeleteNode(nodeID int) error

	// Vendor operations
	GetVendor(vendorID int) (*models.VendorInfo, error)
	GetVendors() ([]*models.VendorInfo, error)
	SaveVendor(vendor *models.VendorInfo) error

	// Settings operations
	GetSetting(key string) (interface{}, error)
	SaveSetting(key string, value interface{}) error
	DeleteSetting(key string) error

	// Lifecycle
	Start() error
	Stop() error
	Sync() error
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(basePath string, log *logger.Logger) *JSONStorage {
	return &JSONStorage{
		basePath: basePath,
		logger:   log,
		nodes:    make(map[int]*models.MatterNodeData),
		vendors:  make(map[int]*models.VendorInfo),
		settings: make(map[string]interface{}),
	}
}

// Start initializes the storage and loads existing data
func (s *JSONStorage) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Load existing data
	if err := s.loadNodes(); err != nil {
		s.logger.Warn("Failed to load nodes", logger.ErrorField(err))
	}

	if err := s.loadVendors(); err != nil {
		s.logger.Warn("Failed to load vendors", logger.ErrorField(err))
	}

	if err := s.loadSettings(); err != nil {
		s.logger.Warn("Failed to load settings", logger.ErrorField(err))
	}

	s.logger.Info("JSON storage started",
		logger.String("path", s.basePath),
		logger.Int("nodes", len(s.nodes)),
		logger.Int("vendors", len(s.vendors)),
	)

	return nil
}

// Stop saves all data and closes storage
func (s *JSONStorage) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.sync(); err != nil {
		return fmt.Errorf("failed to sync data during stop: %w", err)
	}

	s.logger.Info("JSON storage stopped")
	return nil
}

// Sync writes all in-memory data to disk
func (s *JSONStorage) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.sync()
}

func (s *JSONStorage) sync() error {
	if err := s.saveNodes(); err != nil {
		return fmt.Errorf("failed to save nodes: %w", err)
	}

	if err := s.saveVendors(); err != nil {
		return fmt.Errorf("failed to save vendors: %w", err)
	}

	if err := s.saveSettings(); err != nil {
		return fmt.Errorf("failed to save settings: %w", err)
	}

	return nil
}

// Node operations

func (s *JSONStorage) GetNode(nodeID int) (*models.MatterNodeData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %d not found", nodeID)
	}

	// Return a copy to prevent external modification
	nodeCopy := *node
	return &nodeCopy, nil
}

func (s *JSONStorage) GetNodes() ([]*models.MatterNodeData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]*models.MatterNodeData, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodeCopy := *node
		nodes = append(nodes, &nodeCopy)
	}

	return nodes, nil
}

func (s *JSONStorage) SaveNode(node *models.MatterNodeData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Store a copy to prevent external modification
	nodeCopy := *node
	s.nodes[node.NodeID] = &nodeCopy

	return s.saveNodes()
}

func (s *JSONStorage) DeleteNode(nodeID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.nodes, nodeID)
	return s.saveNodes()
}

// Vendor operations

func (s *JSONStorage) GetVendor(vendorID int) (*models.VendorInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vendor, exists := s.vendors[vendorID]
	if !exists {
		return nil, fmt.Errorf("vendor %d not found", vendorID)
	}

	vendorCopy := *vendor
	return &vendorCopy, nil
}

func (s *JSONStorage) GetVendors() ([]*models.VendorInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	vendors := make([]*models.VendorInfo, 0, len(s.vendors))
	for _, vendor := range s.vendors {
		vendorCopy := *vendor
		vendors = append(vendors, &vendorCopy)
	}

	return vendors, nil
}

func (s *JSONStorage) SaveVendor(vendor *models.VendorInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	vendorCopy := *vendor
	s.vendors[vendor.VendorID] = &vendorCopy

	return s.saveVendors()
}

// Settings operations

func (s *JSONStorage) GetSetting(key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.settings[key]
	if !exists {
		return nil, fmt.Errorf("setting %s not found", key)
	}

	return value, nil
}

func (s *JSONStorage) SaveSetting(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.settings[key] = value
	return s.saveSettings()
}

func (s *JSONStorage) DeleteSetting(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.settings, key)
	return s.saveSettings()
}

// File operations

func (s *JSONStorage) loadNodes() error {
	path := filepath.Join(s.basePath, "nodes.json")
	return s.loadJSONFile(path, &s.nodes)
}

func (s *JSONStorage) saveNodes() error {
	path := filepath.Join(s.basePath, "nodes.json")
	return s.saveJSONFile(path, s.nodes)
}

func (s *JSONStorage) loadVendors() error {
	path := filepath.Join(s.basePath, "vendors.json")
	return s.loadJSONFile(path, &s.vendors)
}

func (s *JSONStorage) saveVendors() error {
	path := filepath.Join(s.basePath, "vendors.json")
	return s.saveJSONFile(path, s.vendors)
}

func (s *JSONStorage) loadSettings() error {
	path := filepath.Join(s.basePath, "settings.json")
	return s.loadJSONFile(path, &s.settings)
}

func (s *JSONStorage) saveSettings() error {
	path := filepath.Join(s.basePath, "settings.json")
	return s.saveJSONFile(path, s.settings)
}

func (s *JSONStorage) loadJSONFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist, that's OK
		}
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}

	if len(data) == 0 {
		return nil // Empty file, that's OK
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON from %s: %w", path, err)
	}

	return nil
}

func (s *JSONStorage) saveJSONFile(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to temporary file first
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file %s: %w", tmpPath, err)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

// BackupData creates a backup of all stored data
func (s *JSONStorage) BackupData() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupPath := filepath.Join(s.basePath, "backup_"+timestamp)

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy all data files
	files := []string{"nodes.json", "vendors.json", "settings.json"}
	for _, file := range files {
		src := filepath.Join(s.basePath, file)
		dst := filepath.Join(backupPath, file)

		if err := s.copyFile(src, dst); err != nil {
			s.logger.Warn("Failed to backup file", logger.String("file", file), logger.ErrorField(err))
		}
	}

	s.logger.Info("Data backup created", logger.String("path", backupPath))
	return nil
}

func (s *JSONStorage) copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Source doesn't exist, skip
		}
		return err
	}

	return os.WriteFile(dst, data, 0644)
}
