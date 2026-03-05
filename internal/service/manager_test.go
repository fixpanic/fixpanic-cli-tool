package service

import (
	"os"
	"strings"
	"testing"

	"github.com/fixpanic/opssquad-cli-tool/internal/platform"
)

// MockFileSystem
type MockFileSystem struct {
	MkdirAllFunc  func(path string, perm os.FileMode) error
	WriteFileFunc func(name string, data []byte, perm os.FileMode) error
	RemoveFunc    func(name string) error
	StatFunc      func(name string) (os.FileInfo, error)
}

func (m *MockFileSystem) MkdirAll(path string, perm os.FileMode) error { return m.MkdirAllFunc(path, perm) }
func (m *MockFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return m.WriteFileFunc(name, data, perm)
}
func (m *MockFileSystem) Remove(name string) error { return m.RemoveFunc(name) }
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) { return m.StatFunc(name) }

// MockCommandRunner
type MockCommandRunner struct {
	RunFunc            func(name string, arg ...string) error
	OutputFunc         func(name string, arg ...string) ([]byte, error)
	CombinedOutputFunc func(name string, arg ...string) ([]byte, error)
}

func (m *MockCommandRunner) Run(name string, arg ...string) error { return m.RunFunc(name, arg...) }
func (m *MockCommandRunner) Output(name string, arg ...string) ([]byte, error) {
	return m.OutputFunc(name, arg...)
}
func (m *MockCommandRunner) CombinedOutput(name string, arg ...string) ([]byte, error) {
	return m.CombinedOutputFunc(name, arg...)
}

func TestManager_GenerateServiceFile(t *testing.T) {
	tests := []struct {
		name     string
		isRoot   bool
		contains []string
	}{
		{
			name:   "root service",
			isRoot: true,
			contains: []string{
				"User=root",
				"WantedBy=multi-user.target",
			},
		},
		{
			name:   "user service",
			isRoot: false,
			contains: []string{
				"WantedBy=default.target",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &platform.PlatformInfo{
				IsRoot:    tt.isRoot,
				LibDir:    "/usr/local/lib/opssquad",
				ConfigDir: "/etc/opssquad",
			}
			m := &Manager{platform: p}
			got, err := m.generateServiceFile()
			if err != nil {
				t.Fatalf("generateServiceFile failed: %v", err)
			}

			for _, s := range tt.contains {
				if !strings.Contains(got, s) {
					t.Errorf("Expected service file to contain %q", s)
				}
			}

			if !tt.isRoot && strings.Contains(got, "User=root") {
				t.Error("User=root should not be present in user service")
			}
		})
	}
}

func TestManager_Install(t *testing.T) {
	origIsSystemdAvailable := platform.IsSystemdAvailable
	platform.IsSystemdAvailable = func() bool { return true }
	t.Cleanup(func() { platform.IsSystemdAvailable = origIsSystemdAvailable })

	p := &platform.PlatformInfo{LibDir: "/tmp/lib", ConfigDir: "/tmp/etc"}
	fs := &MockFileSystem{
		MkdirAllFunc:  func(path string, perm os.FileMode) error { return nil },
		WriteFileFunc: func(name string, data []byte, perm os.FileMode) error { return nil },
	}
	runner := &MockCommandRunner{
		CombinedOutputFunc: func(name string, arg ...string) ([]byte, error) { return nil, nil },
	}
	m := &Manager{platform: p, fs: fs, runner: runner}
	
	err := m.Install()
	if err != nil {
		t.Errorf("Install failed: %v", err)
	}
}

func TestManager_EnableDisable(t *testing.T) {
	origIsSystemdAvailable := platform.IsSystemdAvailable
	platform.IsSystemdAvailable = func() bool { return true }
	t.Cleanup(func() { platform.IsSystemdAvailable = origIsSystemdAvailable })

	p := &platform.PlatformInfo{}
	runner := &MockCommandRunner{
		CombinedOutputFunc: func(name string, arg ...string) ([]byte, error) { return nil, nil },
		RunFunc:            func(name string, arg ...string) error { return nil },
	}
	m := &Manager{platform: p, runner: runner}

	if err := m.Enable(); err != nil {
		t.Errorf("Enable failed: %v", err)
	}
	if err := m.Disable(); err != nil {
		t.Errorf("Disable failed: %v", err)
	}
}
