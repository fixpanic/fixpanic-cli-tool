package process

import (
	"os"
	"testing"
)

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

func TestBaseProcessManager_IsProcessRunning(t *testing.T) {
	bm := &BaseProcessManager{Runner: &MockCommandRunner{}}
	
	t.Run("invalid pid", func(t *testing.T) {
		if bm.IsProcessRunning(0) {
			t.Error("Expected false for PID 0")
		}
		if bm.IsProcessRunning(-1) {
			t.Error("Expected false for PID -1")
		}
	})

	t.Run("current process", func(t *testing.T) {
		pid := os.Getpid()
		if !bm.IsProcessRunning(pid) {
			t.Errorf("Expected true for current PID %d", pid)
		}
	})
}

func TestBaseProcessManager_GetProcessStatus(t *testing.T) {
	bm := &BaseProcessManager{Runner: &MockCommandRunner{}}
	pid := os.Getpid()
	
	status := bm.GetProcessStatus(pid)
	if status.PID != pid {
		t.Errorf("Expected PID %d, got %d", pid, status.PID)
	}
	if !status.Running {
		t.Error("Expected Running to be true")
	}
}
