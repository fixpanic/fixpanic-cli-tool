package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AgentConfig represents the agent configuration
type AgentConfig struct {
	App        AppSection        `yaml:"app"`
	ReqHandler ReqHandlerSection `yaml:"req_handler"`
	Logging    LoggingSection    `yaml:"logging"`
}

type AppSection struct {
	AgentID                string `yaml:"agent_id"`
	APIKey                 string `yaml:"api_key"`
	TLSEnabled             bool   `yaml:"tls_enabled"`
	TLSInsecureSkipVerify  bool   `yaml:"tls_insecure_skip_verify"`
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

// DefaultConfig returns a default configuration with TLS enabled
func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		App: AppSection{
			TLSEnabled:            true,  // Enable TLS by default for security
			TLSInsecureSkipVerify: false, // Require valid certificates
		},
		ReqHandler: ReqHandlerSection{
			MaxConcurrentConnections: 10,
			ConnectionTimeout:        "60s",
			DefaultToolTimeout:       300,
			TLSEnabled:               true,  // Enable TLS by default for security
			TLSInsecureSkipVerify:    false, // Require valid certificates
		},
		Logging: LoggingSection{
			Level: "info",
			File:  "/var/log/fixpanic/agent.log",
		},
	}
}

// LoadConfig loads configuration from file
func LoadConfig(path string) (*AgentConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AgentConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *AgentConfig, path string) error {
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
func (c *AgentConfig) Validate() error {
	if c.App.AgentID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if c.App.APIKey == "" {
		return fmt.Errorf("agent API key is required")
	}
	return nil
}

// GetConfigPath returns the default config path
func GetConfigPath() string {
	return "/etc/fixpanic/agent.yaml"
}

// GetUserConfigPath returns the user-specific config path
func GetUserConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".fixpanic.yaml"
	}
	return filepath.Join(home, ".fixpanic", "agent.yaml")
}
