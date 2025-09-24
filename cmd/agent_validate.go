package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/fixpanic/fixpanic-cli/internal/config"
	"github.com/fixpanic/fixpanic-cli/internal/connectivity"
	"github.com/fixpanic/fixpanic-cli/internal/platform"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// agentValidateCmd represents the agent validate-rules command
var agentValidateCmd = &cobra.Command{
	Use:   "validate-rules",
	Short: "Validate security rules configuration",
	Long: `Validate the security rules configuration for the Fixpanic agent.
	
This command checks if the security rules file exists and is properly formatted.
Security rules define which commands the agent is allowed to execute.`,
	Example: `  # Validate security rules
  fixpanic agent validate-rules`,
	RunE: runAgentValidate,
}

func init() {
	agentCmd.AddCommand(agentValidateCmd)
}

func runAgentValidate(cmd *cobra.Command, args []string) error {
	fmt.Println("Validating security rules configuration...")

	// Get platform information
	platformInfo, err := platform.GetPlatformInfo()
	if err != nil {
		return fmt.Errorf("failed to get platform info: %w", err)
	}

	// Check if connectivity layer is installed
	connectivityManager := connectivity.NewManager(platformInfo)
	if !connectivityManager.IsInstalled() {
		return fmt.Errorf("agent is not installed. Run 'fixpanic agent install' first")
	}

	// Load configuration
	configPath := platformInfo.GetConfigPath()
	agentConfig, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if security rules file is configured
	rulesFile := agentConfig.Security.RulesFile
	if rulesFile == "" {
		rulesFile = "/etc/fixpanic/security-rules.yaml"
	}

	fmt.Printf("Checking security rules file: %s\n", rulesFile)

	// Check if rules file exists
	if _, err := os.Stat(rulesFile); os.IsNotExist(err) {
		fmt.Println("⚠️  Security rules file does not exist")
		fmt.Println("A default rules file will be created with basic restrictions.")

		// Create default rules file
		if err := createDefaultRulesFile(rulesFile); err != nil {
			return fmt.Errorf("failed to create default rules file: %w", err)
		}

		fmt.Printf("✅ Created default security rules file: %s\n", rulesFile)
		return nil
	}

	// Read and validate rules file
	rulesContent, err := os.ReadFile(rulesFile)
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	// Parse YAML
	var rules SecurityRules
	if err := yaml.Unmarshal(rulesContent, &rules); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// Validate rules structure
	if err := validateRulesStructure(&rules); err != nil {
		return fmt.Errorf("invalid rules structure: %w", err)
	}

	fmt.Println("✅ Security rules file is valid")
	fmt.Printf("   Rules version: %s\n", rules.Version)
	fmt.Printf("   Allow patterns: %d\n", len(rules.Allow))
	fmt.Printf("   Deny patterns: %d\n", len(rules.Deny))

	if len(rules.Allow) > 0 {
		fmt.Println("   Allowed commands:")
		for _, pattern := range rules.Allow {
			fmt.Printf("     - %s\n", pattern)
		}
	}

	if len(rules.Deny) > 0 {
		fmt.Println("   Denied commands:")
		for _, pattern := range rules.Deny {
			fmt.Printf("     - %s\n", pattern)
		}
	}

	// Test rule matching
	fmt.Println("\nTesting rule matching...")
	testCommands := []string{
		"ls -la",
		"cat /etc/passwd",
		"rm -rf /",
		"ps aux",
		"systemctl restart nginx",
	}

	for _, cmd := range testCommands {
		allowed := isCommandAllowed(&rules, cmd)
		status := "✅ ALLOWED"
		if !allowed {
			status = "❌ DENIED"
		}
		fmt.Printf("   %s: %s\n", cmd, status)
	}

	fmt.Println("\n✅ Security rules validation completed successfully!")
	return nil
}

// SecurityRules represents the security rules configuration
type SecurityRules struct {
	Version string   `yaml:"version"`
	Allow   []string `yaml:"allow"`
	Deny    []string `yaml:"deny"`
}

// createDefaultRulesFile creates a default security rules file
func createDefaultRulesFile(path string) error {
	defaultRules := `# Fixpanic Security Rules
# This file defines which commands the agent is allowed to execute
# 
# Format:
# - allow: list of allowed command patterns (regex)
# - deny: list of denied command patterns (regex)
# 
# Rules are evaluated in order: deny rules take precedence over allow rules

version: "1.0"

# Allowed commands - basic system monitoring and diagnostics
allow:
  - "^ls\\s"                    # List directory contents
  - "^ps\\s"                     # Process status
  - "^top\\s"                    # System monitor
  - "^df\\s"                     # Disk space
  - "^free\\s"                   # Memory usage
  - "^uptime\\s"                 # System uptime
  - "^uname\\s"                   # System information
  - "^cat\\s"                     # Read files (use with caution)
  - "^grep\\s"                    # Search in files
  - "^find\\s"                    # Find files
  - "^netstat\\s"                 # Network connections
  - "^ss\\s"                      # Socket statistics
  - "^ping\\s"                    # Network connectivity test
  - "^traceroute\\s"              # Network route tracing
  - "^journalctl\\s"              # System logs
  - "^systemctl\\sstatus\\s"       # Service status
  - "^docker\\sps\\s"              # Docker containers
  - "^docker\\slogs\\s"            # Docker logs
  - "^kubectl\\sget\\s"            # Kubernetes resources
  - "^kubectl\\slogs\\s"            # Kubernetes logs

# Denied commands - potentially dangerous operations
deny:
  - "^rm\\s.*-rf\\s"              # Recursive force delete
  - "^dd\\s"                      # Disk operations
  - "^mkfs\\s"                    # File system creation
  - "^fdisk\\s"                    # Disk partitioning
  - "^parted\\s"                   # Partition editor
  - "^chmod\\s.*777\\s"            # World-writable permissions
  - "^chown\\s.*root\\s"           # Root ownership changes
  - "^sudo\\s"                     # Privilege escalation
  - "^su\\s"                       # Switch user
  - "^passwd\\s"                   # Password changes
  - "^useradd\\s"                  # User addition
  - "^userdel\\s"                  # User deletion
  - "^systemctl\\s(restart|stop|start)\\s"  # Service control
  - "^reboot\\s"                   # System reboot
  - "^shutdown\\s"                 # System shutdown
  - "^halt\\s"                     # System halt
  - "^poweroff\\s"                  # System poweroff
  - "^curl.*\\|\\s*sh\\s"           # Pipe to shell (security risk)
  - "^wget.*\\|\\s*sh\\s"           # Pipe to shell (security risk)
`

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write the rules file
	if err := os.WriteFile(path, []byte(defaultRules), 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	return nil
}

// validateRulesStructure validates the security rules structure
func validateRulesStructure(rules *SecurityRules) error {
	if rules.Version == "" {
		return fmt.Errorf("rules version is required")
	}

	// Validate allow patterns
	for i, pattern := range rules.Allow {
		if pattern == "" {
			return fmt.Errorf("empty allow pattern at index %d", i)
		}
		// Basic regex validation
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex in allow pattern '%s': %w", pattern, err)
		}
	}

	// Validate deny patterns
	for i, pattern := range rules.Deny {
		if pattern == "" {
			return fmt.Errorf("empty deny pattern at index %d", i)
		}
		// Basic regex validation
		if _, err := regexp.Compile(pattern); err != nil {
			return fmt.Errorf("invalid regex in deny pattern '%s': %w", pattern, err)
		}
	}

	return nil
}

// isCommandAllowed checks if a command is allowed based on the rules
func isCommandAllowed(rules *SecurityRules, command string) bool {
	// Check deny patterns first (deny takes precedence)
	for _, pattern := range rules.Deny {
		if matched, _ := regexp.MatchString(pattern, command); matched {
			return false
		}
	}

	// Check allow patterns
	for _, pattern := range rules.Allow {
		if matched, _ := regexp.MatchString(pattern, command); matched {
			return true
		}
	}

	// Default: deny if no allow patterns match
	return false
}
