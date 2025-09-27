# Fixpanic Agent Configuration Reading Implementation Guide

## Overview
You are developing the **fixpanic-agent** (connectivity layer) binary that must read configuration files created by the Fixpanic CLI tool. This guide explains exactly how to implement configuration reading to work seamlessly with the CLI.

## CLI Integration Context

### How the CLI Invokes Your Binary
The CLI starts your binary with these exact command-line arguments:

```bash
# Standard execution
./connectivity --config /etc/fixpanic/agent.yaml

# With optional log level
./connectivity --config /etc/fixpanic/agent.yaml --log-level debug

# Version check
./connectivity --version
```

### Configuration File Locations
The CLI creates configuration files at platform-specific paths:
- **Root user**: `/etc/fixpanic/agent.yaml`
- **Non-root user**: `~/.config/fixpanic/agent.yaml`

## Required Implementation

### 1. Command Line Argument Parsing
Your binary **must** implement a command-line interface that accepts:

```bash
./connectivity --config <config-file-path> [--log-level <level>] [--version]
```

**Implementation Requirements:**
- Parse `--config` flag as required parameter
- Parse `--log-level` flag as optional parameter (default: "info")
- Parse `--version` flag to show binary version
- Exit gracefully with usage information if `--config` is missing
- Handle unknown flags gracefully

### 2. Configuration File Format
The CLI creates YAML configuration files with this exact structure:

```yaml
agent:
  id: "agent_123"           # Your agent identifier
  api_key: "fp_abc123xyz"   # Your API authentication key
logging:
  level: "info"             # Log level (debug, info, warn, error)
  file: "/var/log/fixpanic/agent.log"  # Log file path
```

### 3. Configuration Reading Implementation

#### Step 1: Define Configuration Structure
```go
type AgentConfig struct {
    Agent   AgentSection   `yaml:"agent"`
    Logging LoggingSection `yaml:"logging"`
}

type AgentSection struct {
    ID     string `yaml:"id"`
    APIKey string `yaml:"api_key"`
}

type LoggingSection struct {
    Level string `yaml:"level"`
    File  string `yaml:"file"`
}
```

#### Step 2: Implement Configuration Loading
```go
func LoadConfig(configPath string) (*AgentConfig, error) {
    // Read the configuration file
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
    }

    // Parse YAML
    var config AgentConfig
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
    }

    return &config, nil
}
```

#### Step 3: Validate Configuration
```go
func (c *AgentConfig) Validate() error {
    if c.Agent.ID == "" {
        return fmt.Errorf("agent ID is required")
    }
    if c.Agent.APIKey == "" {
        return fmt.Errorf("agent API key is required")
    }
    if c.Logging.Level == "" {
        c.Logging.Level = "info" // Set default
    }
    if c.Logging.File == "" {
        return fmt.Errorf("log file path is required")
    }
    return nil
}
```

### 4. Main Application Flow

#### Step 1: Parse Command Line Arguments
```go
func main() {
    var configPath string
    var logLevel string

    flag.StringVar(&configPath, "config", "", "Path to configuration file (required)")
    flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
    flag.Parse()

    // Handle version flag
    if len(os.Args) > 1 && os.Args[1] == "--version" {
        fmt.Println("fixpanic-agent v1.0.0")
        return
    }

    // Validate required config path
    if configPath == "" {
        fmt.Fprintf(os.Stderr, "Error: --config flag is required\n")
        fmt.Fprintf(os.Stderr, "Usage: %s --config <config-file>\n", os.Args[0])
        os.Exit(1)
    }

    // Load and validate configuration
    config, err := LoadConfig(configPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Configuration error: %v\n", err)
        os.Exit(1)
    }

    if err := config.Validate(); err != nil {
        fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
        os.Exit(1)
    }

    // Use the configuration
    runAgent(config)
}
```

#### Step 2: Initialize Logging
```go
func setupLogging(config *AgentConfig) error {
    // Create log directory if it doesn't exist
    logDir := filepath.Dir(config.Logging.File)
    if err := os.MkdirAll(logDir, 0755); err != nil {
        return fmt.Errorf("failed to create log directory: %w", err)
    }

    // Configure your logger with config.Logging.Level and config.Logging.File
    // This depends on your logging framework (logrus, zap, etc.)
    return nil
}
```

#### Step 3: Use Agent Credentials
```go
func runAgent(config *AgentConfig) error {
    // Extract credentials
    agentID := config.Agent.ID
    apiKey := config.Agent.APIKey

    // Use these credentials to:
    // 1. Authenticate with Fixpanic platform
    // 2. Register the agent
    // 3. Establish WebSocket connection to socket.fixpanic.com:8080
    // 4. Handle real-time security rules from dashboard

    fmt.Printf("Starting agent %s with API key %s\n", agentID, maskAPIKey(apiKey))

    // Your agent implementation here...
    return nil
}
```

### 5. Error Handling Requirements

#### Configuration File Errors
- **File not found**: Provide clear error message with expected path
- **Permission denied**: Explain file permission requirements
- **Invalid YAML**: Show line number and parsing error
- **Missing fields**: Specify which required fields are missing

#### Runtime Errors
- **Invalid credentials**: Handle authentication failures gracefully
- **Network issues**: Implement retry logic with exponential backoff
- **Configuration changes**: Handle hot-reloading if needed

### 6. Security Considerations

#### File Permissions
- Configuration file has 0600 permissions (owner read/write only)
- Ensure your binary respects these permissions
- Don't log sensitive information (API keys, etc.)

#### Credential Handling
- Never log API keys in plain text
- Store credentials securely in memory
- Clear sensitive data when shutting down

### 7. Integration Testing

#### Test with CLI-Generated Config
```bash
# 1. Install via CLI
fixpanic agent install --agent-id="test_123" --api-key="test_key"

# 2. Test your binary directly
./connectivity --config /etc/fixpanic/agent.yaml

# 3. Test with custom log level
./connectivity --config /etc/fixpanic/agent.yaml --log-level debug
```

#### Expected Behavior
- ✅ Reads configuration file successfully
- ✅ Validates all required fields
- ✅ Uses agent ID and API key for authentication
- ✅ Configures logging according to specification
- ✅ Handles missing or invalid configuration gracefully

## Implementation Checklist

- [ ] Implement command-line argument parsing
- [ ] Add YAML configuration file reading
- [ ] Implement configuration validation
- [ ] Set up logging based on configuration
- [ ] Extract and use agent credentials
- [ ] Handle all error cases gracefully
- [ ] Test integration with CLI-generated configs
- [ ] Add proper security measures for credentials
- [ ] Implement graceful shutdown handling

## Key Integration Points

1. **Configuration Path**: Always expect `--config` flag with file path
2. **YAML Structure**: Match the exact structure provided by CLI
3. **Validation**: Implement the same validation logic as CLI
4. **Error Messages**: Provide clear, actionable error messages
5. **Logging**: Use the logging configuration from the file

This implementation ensures your connectivity layer binary integrates seamlessly with the Fixpanic CLI tool and can read any configuration file it generates.