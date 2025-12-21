package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
	"gopkg.in/yaml.v3"
)

// NodeConfig represents the node configuration
type NodeConfig struct {
	App        AppSection        `yaml:"app"`
	ReqHandler ReqHandlerSection `yaml:"req_handler"`
	Logging    LoggingSection    `yaml:"logging"`
}

type AppSection struct {
	NodeID                string `yaml:"node_id"`
	APIKey                string `yaml:"api_key"`
	TLSEnabled            bool   `yaml:"tls_enabled"`
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify"`
}

type ReqHandlerSection struct {
	MaxConcurrentConnections int    `yaml:"max_concurrent_connections"`
	ConnectionTimeout        string `yaml:"connection_timeout"`
	DefaultToolTimeout       int    `yaml:"default_tool_timeout"`
	TLSEnabled               bool   `yaml:"tls_enabled"`
	TLSInsecureSkipVerify    bool   `yaml:"tls_insecure_skip_verify"`
}

type LoggingSection struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

type DefaultConfigOptions struct {
	TLSEnabled bool
	NodeID     string
	APIKey     string
}

// DefaultConfig returns a default configuration with TLS enabled
func DefaultConfig(options DefaultConfigOptions) *NodeConfig {
	return &NodeConfig{
		App: AppSection{
			TLSEnabled:            options.TLSEnabled, // Enable TLS by default for security
			TLSInsecureSkipVerify: false,              // Require valid certificates
			NodeID:                options.NodeID,
			APIKey:                options.APIKey,
		},
		ReqHandler: ReqHandlerSection{
			MaxConcurrentConnections: 10,
			ConnectionTimeout:        "60s",
			DefaultToolTimeout:       300,
			TLSEnabled:               options.TLSEnabled, // Enable TLS by default for security
			TLSInsecureSkipVerify:    false,              // Require valid certificates
		},
		Logging: LoggingSection{
			Level: "info",
			File:  getLogPath(),
		},
	}
}

func getLogPath() string {
	info, err := platform.GetPlatformInfo()
	if err != nil {
		return "/var/log/opssquad/node.log" // Fallback
	}
	return filepath.Join(info.LogDir, "node.log")
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*NodeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config NodeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *NodeConfig, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *NodeConfig) Validate() error {
	if c.App.NodeID == "" {
		return fmt.Errorf("node ID is required")
	}
	if c.App.APIKey == "" {
		return fmt.Errorf("node API key is required")
	}
	return nil
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	return "/etc/opssquad/node.yaml"
}

// GetUserConfigPath returns the user-specific config path
func GetUserConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".opssquad.yaml"
	}
	return filepath.Join(home, ".opssquad", "node.yaml")
}
