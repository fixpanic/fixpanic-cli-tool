package service

import (
	"os"
	"os/exec"
)

// FileSystem interface for mocking os operations
type FileSystem interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(name string, data []byte, perm os.FileMode) error
	Remove(name string) error
	Stat(name string) (os.FileInfo, error)
}

// CommandRunner interface for mocking exec.Command
type CommandRunner interface {
	Run(name string, arg ...string) error
	Output(name string, arg ...string) ([]byte, error)
	CombinedOutput(name string, arg ...string) ([]byte, error)
}

// RealFileSystem implements FileSystem using os package
type RealFileSystem struct{}

func (f *RealFileSystem) MkdirAll(path string, perm os.FileMode) error       { return os.MkdirAll(path, perm) }
func (f *RealFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}
func (f *RealFileSystem) Remove(name string) error { return os.Remove(name) }
func (f *RealFileSystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }

// RealCommandRunner implements CommandRunner using os/exec
type RealCommandRunner struct{}

func (r *RealCommandRunner) Run(name string, arg ...string) error {
	return exec.Command(name, arg...).Run()
}

func (r *RealCommandRunner) Output(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).Output()
}

func (r *RealCommandRunner) CombinedOutput(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).CombinedOutput()
}
