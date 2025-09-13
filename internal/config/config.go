package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Storage   StorageConfig   `mapstructure:"storage"`
	Matter    MatterConfig    `mapstructure:"matter"`
	Network   NetworkConfig   `mapstructure:"network"`
	Bluetooth BluetoothConfig `mapstructure:"bluetooth"`
	OTA       OTAConfig       `mapstructure:"ota"`
	MDNS      MDNSConfig      `mapstructure:"mdns"`
	Log       LogConfig       `mapstructure:"log"`
}

type ServerConfig struct {
	Port            int      `mapstructure:"port"`
	ListenAddresses []string `mapstructure:"listen_addresses"`
	ServeStatic     bool     `mapstructure:"serve_static"`
}

type StorageConfig struct {
	Path string `mapstructure:"path"`
}

type MatterConfig struct {
	VendorID                  int    `mapstructure:"vendor_id"`
	FabricID                  int    `mapstructure:"fabric_id"`
	PAARoot                   string `mapstructure:"paa_root_cert_dir"`
	EnableTestNetDCL          bool   `mapstructure:"enable_test_net_dcl"`
	DisableServerInteractions bool   `mapstructure:"disable_server_interactions"`
}

type NetworkConfig struct {
	PrimaryInterface string `mapstructure:"primary_interface"`
}

type BluetoothConfig struct {
	AdapterID int  `mapstructure:"adapter_id"`
	Enabled   bool `mapstructure:"enabled"`
}

type OTAConfig struct {
	ProviderDir string `mapstructure:"provider_dir"`
}

type MDNSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Hostname string `mapstructure:"hostname"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(cmd *cobra.Command) (*Config, error) {
	v := viper.New()

	// Load environment file if specified
	envFile, _ := cmd.Flags().GetString("env-file")
	if envFile != "" {
		if err := loadEnvFile(envFile); err != nil {
			return nil, fmt.Errorf("failed to load env file %s: %w", envFile, err)
		}
	} else {
		// Try to load .env from current directory if it exists
		if _, err := os.Stat(".env"); err == nil {
			if err := loadEnvFile(".env"); err != nil {
				// Don't fail if .env exists but can't be loaded, just warn
				// We could add logging here if needed
			}
		}
	}

	// Set defaults
	setDefaults(v)

	// Read from config file
	configFile, _ := cmd.Flags().GetString("config")
	if configFile != "" {
		v.SetConfigFile(configFile)
	} else {
		// Look for config in home directory
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}

		v.AddConfigPath(filepath.Join(home, ".matter_server"))
		v.AddConfigPath(".")
		v.SetConfigType("yaml")
		v.SetConfigName("config")
	}

	// Read config file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Environment variables
	v.SetEnvPrefix("MATTER")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()

	// Bind command line flags
	if err := bindFlags(cmd, v); err != nil {
		return nil, fmt.Errorf("failed to bind flags: %w", err)
	}

	// Unmarshal config
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set default storage path if not provided
	if cfg.Storage.Path == "" {
		// Use current working directory as default instead of home directory
		pwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current working directory: %w", err)
		}
		cfg.Storage.Path = filepath.Join(pwd, ".matter_server")
	}

	// Validate config
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", 5580)
	v.SetDefault("matter.vendor_id", 0xFFF1)
	v.SetDefault("matter.fabric_id", 1)
	v.SetDefault("matter.enable_test_net_dcl", false)
	v.SetDefault("matter.disable_server_interactions", false)
	v.SetDefault("bluetooth.adapter_id", -1)
	v.SetDefault("bluetooth.enabled", false)
	v.SetDefault("mdns.enabled", true)
	v.SetDefault("mdns.hostname", getDefaultHostname())
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) error {
	flags := map[string]string{
		"port":                        "server.port",
		"listen":                      "server.listen_addresses",
		"storage-path":                "storage.path",
		"vendor-id":                   "matter.vendor_id",
		"fabric-id":                   "matter.fabric_id",
		"primary-interface":           "network.primary_interface",
		"paa-root-cert-dir":           "matter.paa_root_cert_dir",
		"enable-test-net-dcl":         "matter.enable_test_net_dcl",
		"bluetooth-adapter":           "bluetooth.adapter_id",
		"ota-provider-dir":            "ota.provider_dir",
		"disable-server-interactions": "matter.disable_server_interactions",
		"mdns-enabled":                "mdns.enabled",
		"mdns-hostname":               "mdns.hostname",
		"log-level":                   "log.level",
		"log-format":                  "log.format",
	}

	for flag, key := range flags {
		if err := v.BindPFlag(key, cmd.Flags().Lookup(flag)); err != nil {
			return fmt.Errorf("failed to bind flag %s: %w", flag, err)
		}
	}

	return nil
}

func validate(cfg *Config) error {
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid port: %d", cfg.Server.Port)
	}

	if cfg.Matter.VendorID < 0 || cfg.Matter.VendorID > 0xFFFF {
		return fmt.Errorf("invalid vendor ID: %d", cfg.Matter.VendorID)
	}

	if cfg.Matter.FabricID <= 0 {
		return fmt.Errorf("invalid fabric ID: %d", cfg.Matter.FabricID)
	}

	return nil
}

func getDefaultHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "matter-server"
	}
	return hostname
}
