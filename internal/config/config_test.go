package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	options := DefaultConfigOptions{
		TLSEnabled: true,
		NodeID:     "test-node",
		Token:      "test-token",
	}

	cfg := DefaultConfig(options)

	if cfg.App.NodeID != options.NodeID {
		t.Errorf("Expected NodeID %s, got %s", options.NodeID, cfg.App.NodeID)
	}
	if cfg.App.Token != options.Token {
		t.Errorf("Expected Token %s, got %s", options.Token, cfg.App.Token)
	}
	if cfg.App.TLSEnabled != options.TLSEnabled {
		t.Errorf("Expected TLSEnabled %v, got %v", options.TLSEnabled, cfg.App.TLSEnabled)
	}
	if cfg.ReqHandler.TLSEnabled != options.TLSEnabled {
		t.Errorf("Expected ReqHandler TLSEnabled %v, got %v", options.TLSEnabled, cfg.ReqHandler.TLSEnabled)
	}
}

func TestNodeConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *NodeConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &NodeConfig{
				App: AppSection{
					NodeID: "node1",
					Token:  "token1",
				},
			},
			wantErr: false,
		},
		{
			name: "missing node id",
			config: &NodeConfig{
				App: AppSection{
					Token: "token1",
				},
			},
			wantErr: true,
		},
		{
			name: "missing token",
			config: &NodeConfig{
				App: AppSection{
					NodeID: "node1",
				},
			},
			wantErr: true,
		},
		{
			name: "empty config",
			config: &NodeConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "node.yaml")

	originalCfg := &NodeConfig{
		App: AppSection{
			NodeID:     "test-node",
			Token:      "test-token",
			TLSEnabled: true,
		},
		ReqHandler: ReqHandlerSection{
			MaxConcurrentConnections: 5,
			ConnectionTimeout:        "30s",
		},
		Logging: LoggingSection{
			Level: "debug",
			File:  "/tmp/test.log",
		},
	}

	// Test Save
	err := SaveConfig(originalCfg, configPath)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("Config file was not created at %s", configPath)
	}

	// Test Load
	loadedCfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Compare fields
	if loadedCfg.App.NodeID != originalCfg.App.NodeID {
		t.Errorf("Loaded NodeID %s, want %s", loadedCfg.App.NodeID, originalCfg.App.NodeID)
	}
	if loadedCfg.Logging.Level != originalCfg.Logging.Level {
		t.Errorf("Loaded Logging Level %s, want %s", loadedCfg.Logging.Level, originalCfg.Logging.Level)
	}
	if loadedCfg.ReqHandler.MaxConcurrentConnections != originalCfg.ReqHandler.MaxConcurrentConnections {
		t.Errorf("Loaded MaxConcurrentConnections %d, want %d", loadedCfg.ReqHandler.MaxConcurrentConnections, originalCfg.ReqHandler.MaxConcurrentConnections)
	}
}
