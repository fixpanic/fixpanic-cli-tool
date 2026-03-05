package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestShouldUseColors(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected bool
	}{
		{
			name:     "no color env",
			env:      map[string]string{"NO_COLOR": "1"},
			expected: false,
		},
		{
			name:     "clicolor 0",
			env:      map[string]string{"CLICOLOR": "0"},
			expected: false,
		},
		{
			name:     "default colors enabled",
			env:      map[string]string{"NO_COLOR": "", "CLICOLOR": "1"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear relevant env vars
			os.Unsetenv("NO_COLOR")
			os.Unsetenv("CLICOLOR")
			os.Unsetenv("FORCE_COLOR")
			os.Unsetenv("CLICOLOR_FORCE")

			for k, v := range tt.env {
				if v != "" {
					os.Setenv(k, v)
				}
			}

			got := shouldUseColors()
			if got != tt.expected {
				t.Errorf("shouldUseColors() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestLoggerOutputs(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	l := &Logger{useColors: false}
	
	l.Info("test info")
	l.Success("test success")
	l.Warning("test warning")
	l.Error("test error")
	l.Progress("test progress")
	l.Step(1, "test step")
	l.KeyValue("Key", "Value")
	l.List("test list")
	l.Command("ls")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	expectedSubstrings := []string{
		"[INFO] test info",
		"[SUCCESS] test success",
		"[WARNING] test warning",
		"[ERROR] test error",
		"[PROGRESS] test progress",
		"[STEP 1] test step",
		"Key: Value",
		"✓ test list",
		"$ ls",
	}

	for _, s := range expectedSubstrings {
		if !strings.Contains(output, s) {
			t.Errorf("Output missing expected string: %q", s)
		}
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Simple smoke test to ensure package-level functions don't panic
	Info("hello %s", "world")
	Success("done")
	Warning("careful")
	Error("failed")
	Progress("working")
	Step(1, "first")
	Header("Testing")
	Separator()
	KeyValue("a", "b")
	List("item")
}
