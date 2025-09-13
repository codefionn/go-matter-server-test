package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestSetDefaults(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected interface{}
	}{
		{"Server Port", "server.port", 5580},
		{"Vendor ID", "matter.vendor_id", 0xFFF1},
		{"Fabric ID", "matter.fabric_id", 1},
		{"Enable Test Net DCL", "matter.enable_test_net_dcl", false},
		{"Disable Server Interactions", "matter.disable_server_interactions", false},
		{"Bluetooth Adapter ID", "bluetooth.adapter_id", -1},
		{"Bluetooth Enabled", "bluetooth.enabled", false},
		{"Log Level", "log.level", "info"},
		{"Log Format", "log.format", "console"},
	}

	// Create a viper instance and set defaults
	v := createTestViper()
	setDefaults(v)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := v.Get(tt.key)
			if actual != tt.expected {
				t.Errorf("Expected %s to be %v, got %v", tt.key, tt.expected, actual)
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		expectErr bool
	}{
		{
			name: "Valid config",
			config: &Config{
				Server: ServerConfig{
					Port: 5580,
				},
				Matter: MatterConfig{
					VendorID: 0xFFF1,
					FabricID: 1,
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid port - negative",
			config: &Config{
				Server: ServerConfig{
					Port: -1,
				},
				Matter: MatterConfig{
					VendorID: 0xFFF1,
					FabricID: 1,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid port - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 70000,
				},
				Matter: MatterConfig{
					VendorID: 0xFFF1,
					FabricID: 1,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid vendor ID - negative",
			config: &Config{
				Server: ServerConfig{
					Port: 5580,
				},
				Matter: MatterConfig{
					VendorID: -1,
					FabricID: 1,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid vendor ID - too high",
			config: &Config{
				Server: ServerConfig{
					Port: 5580,
				},
				Matter: MatterConfig{
					VendorID: 0x10000,
					FabricID: 1,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid fabric ID - zero",
			config: &Config{
				Server: ServerConfig{
					Port: 5580,
				},
				Matter: MatterConfig{
					VendorID: 0xFFF1,
					FabricID: 0,
				},
			},
			expectErr: true,
		},
		{
			name: "Invalid fabric ID - negative",
			config: &Config{
				Server: ServerConfig{
					Port: 5580,
				},
				Matter: MatterConfig{
					VendorID: 0xFFF1,
					FabricID: -1,
				},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if tt.expectErr && err == nil {
				t.Error("Expected validation error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "test_config.yaml")

	configContent := `
server:
  port: 8080
  listen_addresses: ["127.0.0.1", "::1"]
  serve_static: true

storage:
  path: "/test/storage"

matter:
  vendor_id: 0x1234
  fabric_id: 5
  enable_test_net_dcl: true
  disable_server_interactions: true

network:
  primary_interface: "eth0"

bluetooth:
  adapter_id: 1
  enabled: true

ota:
  provider_dir: "/test/ota"

log:
  level: "debug"
  format: "json"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Create test command with config file flag
	cmd := &cobra.Command{}
	setupTestFlags(cmd)
	cmd.Flags().Set("config", configFile)

	// Load config
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify config values
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", cfg.Server.Port)
	}
	if len(cfg.Server.ListenAddresses) != 2 {
		t.Errorf("Expected 2 listen addresses, got %d", len(cfg.Server.ListenAddresses))
	}
	if cfg.Storage.Path != "/test/storage" {
		t.Errorf("Expected storage path '/test/storage', got %s", cfg.Storage.Path)
	}
	if cfg.Matter.VendorID != 0x1234 {
		t.Errorf("Expected vendor ID 0x1234, got %d", cfg.Matter.VendorID)
	}
	if cfg.Matter.FabricID != 5 {
		t.Errorf("Expected fabric ID 5, got %d", cfg.Matter.FabricID)
	}
	if !cfg.Matter.EnableTestNetDCL {
		t.Error("Expected EnableTestNetDCL to be true")
	}
	if !cfg.Matter.DisableServerInteractions {
		t.Error("Expected DisableServerInteractions to be true")
	}
	if cfg.Network.PrimaryInterface != "eth0" {
		t.Errorf("Expected primary interface 'eth0', got %s", cfg.Network.PrimaryInterface)
	}
	if cfg.Bluetooth.AdapterID != 1 {
		t.Errorf("Expected bluetooth adapter ID 1, got %d", cfg.Bluetooth.AdapterID)
	}
	if !cfg.Bluetooth.Enabled {
		t.Error("Expected Bluetooth to be enabled")
	}
	if cfg.OTA.ProviderDir != "/test/ota" {
		t.Errorf("Expected OTA provider dir '/test/ota', got %s", cfg.OTA.ProviderDir)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Expected log level 'debug', got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected log format 'json', got %s", cfg.Log.Format)
	}
}

func TestLoadConfigWithoutFile(t *testing.T) {
	// Create test command without config file
	cmd := &cobra.Command{}
	setupTestFlags(cmd)

	// Load config (should use defaults)
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config without file: %v", err)
	}

	// Verify default values
	if cfg.Server.Port != 5580 {
		t.Errorf("Expected default port 5580, got %d", cfg.Server.Port)
	}
	if cfg.Matter.VendorID != 0xFFF1 {
		t.Errorf("Expected default vendor ID 0xFFF1, got %d", cfg.Matter.VendorID)
	}
	if cfg.Matter.FabricID != 1 {
		t.Errorf("Expected default fabric ID 1, got %d", cfg.Matter.FabricID)
	}
	if cfg.Log.Level != "info" {
		t.Errorf("Expected default log level 'info', got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "console" {
		t.Errorf("Expected default log format 'console', got %s", cfg.Log.Format)
	}
}

func TestLoadConfigWithCommandLineFlags(t *testing.T) {
	// Create test command with flags set
	cmd := &cobra.Command{}
	setupTestFlags(cmd)

	// Set some flags
	cmd.Flags().Set("port", "9000")
	cmd.Flags().Set("vendor-id", "4660") // 0x1234
	cmd.Flags().Set("fabric-id", "10")
	cmd.Flags().Set("log-level", "trace")
	cmd.Flags().Set("log-format", "json")

	// Load config
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config with flags: %v", err)
	}

	// Verify flag values override defaults
	if cfg.Server.Port != 9000 {
		t.Errorf("Expected port 9000 from flag, got %d", cfg.Server.Port)
	}
	if cfg.Matter.VendorID != 4660 {
		t.Errorf("Expected vendor ID 4660 from flag, got %d", cfg.Matter.VendorID)
	}
	if cfg.Matter.FabricID != 10 {
		t.Errorf("Expected fabric ID 10 from flag, got %d", cfg.Matter.FabricID)
	}
	if cfg.Log.Level != "trace" {
		t.Errorf("Expected log level 'trace' from flag, got %s", cfg.Log.Level)
	}
	if cfg.Log.Format != "json" {
		t.Errorf("Expected log format 'json' from flag, got %s", cfg.Log.Format)
	}
}

func TestDefaultStoragePath(t *testing.T) {
	// Create test command
	cmd := &cobra.Command{}
	setupTestFlags(cmd)

	// Load config without setting storage path
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify storage path uses current working directory
	expectedSuffix := ".matter_server"
	if !filepath.IsAbs(cfg.Storage.Path) {
		t.Errorf("Expected absolute storage path, got %s", cfg.Storage.Path)
	}
	if !strings.HasSuffix(cfg.Storage.Path, expectedSuffix) {
		t.Errorf("Expected storage path to end with %s, got %s", expectedSuffix, cfg.Storage.Path)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Set environment variables
	envVars := map[string]string{
		"MATTER_SERVER_PORT":      "7777",
		"MATTER_MATTER_VENDOR_ID": "1111",
		"MATTER_LOG_LEVEL":        "error",
	}

	for key, value := range envVars {
		os.Setenv(key, value)
		defer os.Unsetenv(key)
	}

	// Create test command
	cmd := &cobra.Command{}
	setupTestFlags(cmd)

	// Load config
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config with env vars: %v", err)
	}

	// Verify environment variables are used
	if cfg.Server.Port != 7777 {
		t.Errorf("Expected port 7777 from env var, got %d", cfg.Server.Port)
	}
	if cfg.Matter.VendorID != 1111 {
		t.Errorf("Expected vendor ID 1111 from env var, got %d", cfg.Matter.VendorID)
	}
	if cfg.Log.Level != "error" {
		t.Errorf("Expected log level 'error' from env var, got %s", cfg.Log.Level)
	}
}

func TestConfigPrecedence(t *testing.T) {
	// Create temporary config file
	tempDir := t.TempDir()
	configFile := filepath.Join(tempDir, "precedence_test.yaml")

	configContent := `
server:
  port: 6000

matter:
  vendor_id: 2222

log:
  level: "warn"
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	// Set environment variable
	os.Setenv("MATTER_SERVER_PORT", "7000")
	defer os.Unsetenv("MATTER_SERVER_PORT")

	// Create test command with config file and flags
	cmd := &cobra.Command{}
	setupTestFlags(cmd)
	cmd.Flags().Set("config", configFile)
	cmd.Flags().Set("port", "8000") // This should have highest precedence

	// Load config
	cfg, err := Load(cmd)
	if err != nil {
		t.Fatalf("Failed to load config for precedence test: %v", err)
	}

	// Verify precedence: CLI flag > env var > config file > default
	// Port: CLI flag (8000) should win
	if cfg.Server.Port != 8000 {
		t.Errorf("Expected port 8000 from CLI flag, got %d", cfg.Server.Port)
	}

	// VendorID: Config file (2222) should win (no env var or CLI flag set)
	if cfg.Matter.VendorID != 2222 {
		t.Errorf("Expected vendor ID 2222 from config file, got %d", cfg.Matter.VendorID)
	}

	// LogLevel: Config file (warn) should win
	if cfg.Log.Level != "warn" {
		t.Errorf("Expected log level 'warn' from config file, got %s", cfg.Log.Level)
	}
}

// Helper functions for tests

func createTestViper() *viper.Viper {
	v := viper.New()
	return v
}

func setupTestFlags(cmd *cobra.Command) {
	cmd.Flags().String("config", "", "config file")
	cmd.Flags().String("log-level", "info", "log level")
	cmd.Flags().String("log-format", "console", "log format")
	cmd.Flags().IntP("port", "p", 5580, "WebSocket server port")
	cmd.Flags().StringSlice("listen", []string{}, "Listen addresses")
	cmd.Flags().String("storage-path", "", "Storage path for persistent data")
	cmd.Flags().Int("vendor-id", 0xFFF1, "Vendor ID for the Fabric")
	cmd.Flags().Int("fabric-id", 1, "Fabric ID for the Fabric")
	cmd.Flags().String("primary-interface", "", "Primary network interface")
	cmd.Flags().String("paa-root-cert-dir", "", "Directory where PAA root certificates are stored")
	cmd.Flags().Bool("enable-test-net-dcl", false, "Enable PAA root certificates from test-net DCL")
	cmd.Flags().Int("bluetooth-adapter", -1, "Bluetooth adapter ID")
	cmd.Flags().String("ota-provider-dir", "", "Directory for OTA Provider software updates")
	cmd.Flags().Bool("disable-server-interactions", false, "Disable server cluster interactions")
	cmd.Flags().Bool("mdns-enabled", true, "Enable mDNS hostname advertisement")
	cmd.Flags().String("mdns-hostname", "", "Hostname to advertise via mDNS")
}
