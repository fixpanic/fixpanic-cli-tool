package platform

import (
	"strings"
	"testing"
)

func TestNormalizeArch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"x86_64", "amd64"},
		{"AMD64", "amd64"},
		{"arm64", "arm64"},
		{"aarch64", "arm64"},
		{"i386", "386"},
		{"i686", "386"},
		{"armv7", "arm"},
		{"unknown", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeArch(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeArch(%q) = %q; want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGetBinaryDownloadURL(t *testing.T) {
	// Since GetBinaryDownloadURL calls GetNodePlatformInfo which uses runtime.GOOS/GOARCH,
	// it's harder to test for ALL platforms without mocking.
	// But we can check if it returns a valid looking URL for the current platform.

	version := "v1.0.0"
	url, err := GetBinaryDownloadURL(version)
	if err != nil {
		t.Fatalf("GetBinaryDownloadURL failed: %v", err)
	}

	if !strings.Contains(url, "github.com/fixpanic/opssquad-connectivity-layer-release/releases/download/") {
		t.Errorf("URL doesn't contain expected base: %s", url)
	}

	if !strings.Contains(url, version) {
		t.Errorf("URL doesn't contain version %s: %s", version, url)
	}

	// Test latest
	urlLatest, err := GetBinaryDownloadURL("latest")
	if err != nil {
		t.Fatalf("GetBinaryDownloadURL(latest) failed: %v", err)
	}
	if !strings.Contains(urlLatest, "/latest/download/") {
		t.Errorf("Latest URL doesn't contain /latest/download/: %s", urlLatest)
	}
}

func TestGetBinaryName(t *testing.T) {
	name := GetBinaryName()
	if name == "" {
		t.Error("GetBinaryName returned empty string")
	}
	// On windows it should have .exe, but since we are running on darwin/linux in tests, 
	// we just check it's not empty for now.
}
