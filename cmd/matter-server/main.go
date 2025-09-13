package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/codefionn/go-matter-server/internal/config"
	"github.com/codefionn/go-matter-server/internal/logger"
	"github.com/codefionn/go-matter-server/internal/server"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	rootCmd := &cobra.Command{
		Use:     "matter-server",
		Short:   "Go Matter Server - WebSocket-based Matter controller server",
		Version: fmt.Sprintf("%s (%s)", version, commit),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runServer(ctx, cmd)
		},
	}

	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.matter_server/config.yaml)")
	rootCmd.PersistentFlags().String("env-file", "", "env file to load environment variables from (e.g., .env)")
	rootCmd.PersistentFlags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "console", "log format (console, json)")

	// Server specific flags
	rootCmd.Flags().IntP("port", "p", 5580, "WebSocket server port")
	rootCmd.Flags().StringSliceP("listen", "l", []string{}, "Listen addresses (default: all interfaces)")
	rootCmd.Flags().String("storage-path", "", "Storage path for persistent data (default: $HOME/.matter_server)")
	rootCmd.Flags().Int("vendor-id", 0xFFF1, "Vendor ID for the Fabric")
	rootCmd.Flags().Int("fabric-id", 1, "Fabric ID for the Fabric")
	rootCmd.Flags().String("primary-interface", "", "Primary network interface for link-local addresses")
	rootCmd.Flags().String("paa-root-cert-dir", "", "Directory where PAA root certificates are stored")
	rootCmd.Flags().Bool("enable-test-net-dcl", false, "Enable PAA root certificates from test-net DCL")
	rootCmd.Flags().Int("bluetooth-adapter", -1, "Bluetooth adapter ID for direct commissioning support")
	rootCmd.Flags().String("ota-provider-dir", "", "Directory for OTA Provider software updates")
	rootCmd.Flags().Bool("disable-server-interactions", false, "Disable server cluster interactions")
	rootCmd.Flags().Bool("mdns-enabled", true, "Enable mDNS hostname advertisement")
	rootCmd.Flags().String("mdns-hostname", "", "Hostname to advertise via mDNS (default: system hostname)")

	return rootCmd.ExecuteContext(ctx)
}

func runServer(ctx context.Context, cmd *cobra.Command) error {
	cfg, err := config.Load(cmd)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	log, err := setupLogger(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		return fmt.Errorf("failed to setup logger: %w", err)
	}

	srv, err := server.New(cfg, log)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.Run(ctx)
}

func setupLogger(levelStr, formatStr string) (*logger.Logger, error) {
	level, err := logger.ParseLogLevel(levelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	var format logger.LogFormat
	switch formatStr {
	case "console":
		format = logger.ConsoleFormat
	case "json":
		format = logger.JSONFormat
	default:
		return nil, fmt.Errorf("invalid log format: %s", formatStr)
	}

	return logger.New(logger.Config{
		Level:     level,
		Format:    format,
		UseColors: format == logger.ConsoleFormat,
	}), nil
}
