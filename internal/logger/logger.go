package logger

import (
	"fmt"
	"os"
	"runtime"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Red    = "\033[0;31m"
	Green  = "\033[0;32m"
	Yellow = "\033[1;33m"
	Blue   = "\033[0;34m"
	Purple = "\033[0;35m"
	Cyan   = "\033[0;36m"
	Gray   = "\033[0;37m"
	Bold   = "\033[1m"
)

// Logger provides consistent, colored output for CLI operations
type Logger struct {
	useColors bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		useColors: shouldUseColors(),
	}
}

// shouldUseColors determines if colors should be used based on environment
func shouldUseColors() bool {
	// Disable colors on Windows by default (unless explicitly enabled)
	if runtime.GOOS == "windows" {
		if os.Getenv("FORCE_COLOR") != "true" && os.Getenv("CLICOLOR_FORCE") != "1" {
			return false
		}
	}

	// Check if NO_COLOR is set (universal standard)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if CLICOLOR is explicitly disabled
	if os.Getenv("CLICOLOR") == "0" {
		return false
	}

	// Enable colors by default for terminals
	return true
}

// colorize applies color if colors are enabled
func (l *Logger) colorize(color, text string) string {
	if !l.useColors {
		return text
	}
	return color + text + Reset
}

// Info prints an informational message with blue [INFO] prefix
func (l *Logger) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Blue, "[INFO]")
	fmt.Printf("%s %s\n", prefix, message)
}

// Success prints a success message with green [SUCCESS] prefix
func (l *Logger) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Green, "[SUCCESS]")
	fmt.Printf("%s %s\n", prefix, message)
}

// Warning prints a warning message with yellow [WARNING] prefix
func (l *Logger) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Yellow, "[WARNING]")
	fmt.Printf("%s %s\n", prefix, message)
}

// Error prints an error message with red [ERROR] prefix
func (l *Logger) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Red, "[ERROR]")
	fmt.Printf("%s %s\n", prefix, message)
}

// Progress prints a progress message with cyan [PROGRESS] prefix
func (l *Logger) Progress(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Cyan, "[PROGRESS]")
	fmt.Printf("%s %s\n", prefix, message)
}

// Step prints a numbered step with purple prefix
func (l *Logger) Step(step int, format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Purple, fmt.Sprintf("[STEP %d]", step))
	fmt.Printf("%s %s\n", prefix, message)
}

// Plain prints a message without any prefix (but can still be colored)
func (l *Logger) Plain(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s\n", message)
}

// Header prints a section header with separator
func (l *Logger) Header(title string) {
	separator := "=================================="
	if len(title) > len(separator) {
		separator = ""
		for i := 0; i < len(title); i++ {
			separator += "="
		}
	}

	fmt.Printf("%s\n", l.colorize(Bold+Blue, title))
	fmt.Printf("%s\n", l.colorize(Blue, separator))
}

// Separator prints a visual separator
func (l *Logger) Separator() {
	fmt.Println()
}

// KeyValue prints a key-value pair with consistent formatting
func (l *Logger) KeyValue(key, value string) {
	keyColored := l.colorize(Bold, key+":")
	fmt.Printf("   %s %s\n", keyColored, value)
}

// List prints a bulleted list item
func (l *Logger) List(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	bullet := l.colorize(Green, "✓")
	fmt.Printf("   %s %s\n", bullet, message)
}

// Loading prints a loading message (without newline)
func (l *Logger) Loading(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	prefix := l.colorize(Cyan, "[LOADING]")
	fmt.Printf("%s %s", prefix, message)
}

// LoadingDone completes a loading message
func (l *Logger) LoadingDone(format string, args ...interface{}) {
	if format == "" {
		fmt.Printf(" %s\n", l.colorize(Green, "✓"))
	} else {
		message := fmt.Sprintf(format, args...)
		fmt.Printf(" %s %s\n", l.colorize(Green, "✓"), message)
	}
}

// LoadingFailed completes a loading message with failure
func (l *Logger) LoadingFailed(format string, args ...interface{}) {
	if format == "" {
		fmt.Printf(" %s\n", l.colorize(Red, "✗"))
	} else {
		message := fmt.Sprintf(format, args...)
		fmt.Printf(" %s %s\n", l.colorize(Red, "✗"), message)
	}
}

// Command prints a command that's being executed
func (l *Logger) Command(cmd string) {
	cmdColored := l.colorize(Gray, "$ "+cmd)
	fmt.Printf("   %s\n", cmdColored)
}

// Global logger instance for convenience
var defaultLogger = NewLogger()

// Package-level convenience functions
func Info(format string, args ...interface{})     { defaultLogger.Info(format, args...) }
func Success(format string, args ...interface{})  { defaultLogger.Success(format, args...) }
func Warning(format string, args ...interface{})  { defaultLogger.Warning(format, args...) }
func Error(format string, args ...interface{})    { defaultLogger.Error(format, args...) }
func Progress(format string, args ...interface{}) { defaultLogger.Progress(format, args...) }
func Step(step int, format string, args ...interface{}) { defaultLogger.Step(step, format, args...) }
func Plain(format string, args ...interface{})    { defaultLogger.Plain(format, args...) }
func Header(title string)                          { defaultLogger.Header(title) }
func Separator()                                   { defaultLogger.Separator() }
func KeyValue(key, value string)                   { defaultLogger.KeyValue(key, value) }
func List(format string, args ...interface{})     { defaultLogger.List(format, args...) }
func Loading(format string, args ...interface{})  { defaultLogger.Loading(format, args...) }
func LoadingDone(format string, args ...interface{}) { defaultLogger.LoadingDone(format, args...) }
func LoadingFailed(format string, args ...interface{}) { defaultLogger.LoadingFailed(format, args...) }
func Command(cmd string)                           { defaultLogger.Command(cmd) }