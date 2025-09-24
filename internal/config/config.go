package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AgentConfig represents the agent configuration
type AgentConfig struct {
	Agent    AgentSection    `yaml:"agent"`
	Security SecuritySection `yaml:"security"`
	Logging  LoggingSection  `yaml:"logging"`
}

type AgentSection struct {
	ID           string `yaml:"id"`
	APIKey       string `yaml:"api_key"`
	SocketServer string `yaml:"socket_server"`
}

type SecuritySection struct {
	RulesFile string `yaml:"rules_file"`
}

type LoggingSection struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *AgentConfig {
	return &AgentConfig{
		Agent: AgentSection{
			SocketServer: "socket.fixpanic.com:8080",
		},
		Security: SecuritySection{
			RulesFile: "/etc/fixpanic/security-rules.yaml",
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
	if c.Agent.ID == "" {
		return fmt.Errorf("agent ID is required")
	}
	if c.Agent.APIKey == "" {
		return fmt.Errorf("agent API key is required")
	}
	if c.Agent.SocketServer == "" {
		return fmt.Errorf("socket server address is required")
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
