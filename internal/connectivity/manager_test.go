package connectivity

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
)

// MockHTTPClient
type MockHTTPClient struct {
	GetFunc func(url string) (*http.Response, error)
}

func (m *MockHTTPClient) Get(url string) (*http.Response, error) {
	return m.GetFunc(url)
}

// MockFileSystem
type MockFileSystem struct {
	CreateFunc func(name string) (*os.File, error)
	OpenFunc   func(name string) (*os.File, error)
	StatFunc   func(name string) (os.FileInfo, error)
	RemoveFunc func(name string) error
	RenameFunc func(oldpath, newpath string) error
	ChmodFunc  func(name string, mode os.FileMode) error
}

func (m *MockFileSystem) Create(name string) (*os.File, error) { return m.CreateFunc(name) }
func (m *MockFileSystem) Open(name string) (*os.File, error)   { return m.OpenFunc(name) }
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.StatFunc != nil {
		return m.StatFunc(name)
	}
	return nil, os.ErrNotExist
}
func (m *MockFileSystem) Remove(name string) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(name)
	}
	return nil
}
func (m *MockFileSystem) Rename(oldpath, newpath string) error {
	if m.RenameFunc != nil {
		return m.RenameFunc(oldpath, newpath)
	}
	return nil
}
func (m *MockFileSystem) Chmod(name string, mode os.FileMode) error {
	if m.ChmodFunc != nil {
		return m.ChmodFunc(name, mode)
	}
	return nil
}

// MockCommandRunner
type MockCommandRunner struct {
	RunFunc    func(name string, arg ...string) error
	OutputFunc func(name string, arg ...string) ([]byte, error)
}

func (m *MockCommandRunner) Run(name string, arg ...string) error {
	if m.RunFunc != nil {
		return m.RunFunc(name, arg...)
	}
	return nil
}

func (m *MockCommandRunner) Output(name string, arg ...string) ([]byte, error) {
	if m.OutputFunc != nil {
		return m.OutputFunc(name, arg...)
	}
	return nil, nil
}

func TestManager_IsBinaryInstalled(t *testing.T) {
	p := &platform.PlatformInfo{LibDir: "/tmp"}
	
	t.Run("installed", func(t *testing.T) {
		fs := &MockFileSystem{
			StatFunc: func(name string) (os.FileInfo, error) {
				return nil, nil // exists
			},
		}
		m := &Manager{platform: p, fs: fs}
		if !m.IsBinaryInstalled() {
			t.Error("Expected IsBinaryInstalled to be true")
		}
	})

	t.Run("not installed", func(t *testing.T) {
		fs := &MockFileSystem{
			StatFunc: func(name string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
		}
		m := &Manager{platform: p, fs: fs}
		if m.IsBinaryInstalled() {
			t.Error("Expected IsBinaryInstalled to be false")
		}
	})
}

func TestManager_GetBinaryVersion(t *testing.T) {
	p := &platform.PlatformInfo{LibDir: "/tmp"}
	versionOutput := "opssquad-connectivity-layer v1.2.3"

	fs := &MockFileSystem{
		StatFunc: func(name string) (os.FileInfo, error) {
			return nil, nil
		},
	}
	runner := &MockCommandRunner{
		OutputFunc: func(name string, arg ...string) ([]byte, error) {
			return []byte(versionOutput), nil
		},
	}

	m := &Manager{platform: p, fs: fs, runner: runner}
	got, err := m.GetBinaryVersion()
	if err != nil {
		t.Fatalf("GetBinaryVersion failed: %v", err)
	}

	if got != versionOutput {
		t.Errorf("Expected version %q, got %q", versionOutput, got)
	}
}

func TestManager_GetLatestNodeVersion(t *testing.T) {
	p := &platform.PlatformInfo{}
	mockJSON := `{"tag_name": "v2.0.0"}`
	
	client := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(mockJSON)),
			}, nil
		},
	}

	m := &Manager{platform: p, client: client}
	got, err := m.GetLatestNodeVersion()
	if err != nil {
		t.Fatalf("GetLatestNodeVersion failed: %v", err)
	}

	if got != "v2.0.0" {
		t.Errorf("Expected v2.0.0, got %s", got)
	}
}

func TestManager_VerifyChecksum(t *testing.T) {
	p := &platform.PlatformInfo{LibDir: "/tmp"}
	content := []byte("test content")
	// sha256 of "test content" is 6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72
	expectedChecksum := "6ae8a75555209fd6c44157c0aed8016e763ff435a19cf186f76863140143ff72"

	fs := &MockFileSystem{
		OpenFunc: func(name string) (*os.File, error) {
			// This is tricky with *os.File. In a real scenario we'd refactor Open to return io.ReadCloser.
			// For now, let's use a real temp file for this specific test.
			tmpFile, _ := os.CreateTemp("", "checksum-test")
			tmpFile.Write(content)
			tmpFile.Seek(0, 0)
			return tmpFile, nil
		},
	}
	m := &Manager{platform: p, fs: fs}
	err := m.VerifyChecksum(expectedChecksum)
	if err != nil {
		t.Errorf("VerifyChecksum failed: %v", err)
	}
}

func TestManager_EnsureLatestNode_NoUpdate(t *testing.T) {
	p := &platform.PlatformInfo{LibDir: "/tmp"}
	fs := &MockFileSystem{
		StatFunc: func(name string) (os.FileInfo, error) {
			return nil, nil // installed
		},
	}
	runner := &MockCommandRunner{
		OutputFunc: func(name string, arg ...string) ([]byte, error) {
			return []byte("v1.0.0"), nil
		},
	}
	client := &MockHTTPClient{
		GetFunc: func(url string) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"tag_name": "v1.0.0"}`)),
			}, nil
		},
	}

	m := &Manager{platform: p, fs: fs, runner: runner, client: client}
	err := m.EnsureLatestNode()
	if err != nil {
		t.Errorf("EnsureLatestNode failed: %v", err)
	}
}
