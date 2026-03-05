package connectivity

import (
	"net/http"
	"os"
	"os/exec"
)

// HTTPClient interface for mocking http requests
type HTTPClient interface {
	Get(url string) (resp *http.Response, err error)
}

// FileSystem interface for mocking os operations
type FileSystem interface {
	Create(name string) (*os.File, error)
	Open(name string) (*os.File, error)
	Stat(name string) (os.FileInfo, error)
	Remove(name string) error
	Rename(oldpath, newpath string) error
	Chmod(name string, mode os.FileMode) error
}

// CommandRunner interface for mocking exec.Command
type CommandRunner interface {
	Run(name string, arg ...string) error
	Output(name string, arg ...string) ([]byte, error)
}

// RealHTTPClient implements HTTPClient using net/http
type RealHTTPClient struct {
	client *http.Client
}

func (c *RealHTTPClient) Get(url string) (*http.Response, error) {
	return c.client.Get(url)
}

// RealFileSystem implements FileSystem using os package
type RealFileSystem struct{}

func (f *RealFileSystem) Create(name string) (*os.File, error) { return os.Create(name) }
func (f *RealFileSystem) Open(name string) (*os.File, error)   { return os.Open(name) }
func (f *RealFileSystem) Stat(name string) (os.FileInfo, error) { return os.Stat(name) }
func (f *RealFileSystem) Remove(name string) error             { return os.Remove(name) }
func (f *RealFileSystem) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }
func (f *RealFileSystem) Chmod(name string, mode os.FileMode) error { return os.Chmod(name, mode) }

// RealCommandRunner implements CommandRunner using os/exec
type RealCommandRunner struct{}

func (r *RealCommandRunner) Run(name string, arg ...string) error {
	return exec.Command(name, arg...).Run()
}

func (r *RealCommandRunner) Output(name string, arg ...string) ([]byte, error) {
	return exec.Command(name, arg...).Output()
}
