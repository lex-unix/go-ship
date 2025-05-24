package txman

import (
	"context"
	"fmt"
	"time"

	"neite.dev/go-ship/internal/exec/sshexec"
)

// SSHServiceStub is a stub implementation of the Service interface.
type SSHServiceStub struct {
	hostName      string
	RunFunc       func(ctx context.Context, cmd string, options ...sshexec.SessionOption) error
	WriteFileFunc func(path string, data []byte) error
	ReadFileFunc  func(path string) ([]byte, error)
}

// NewMockSSHLikeService creates a new mock service.
func NewMockSSHLikeService(hostName string) *SSHServiceStub {
	return &SSHServiceStub{
		hostName: hostName,
	}
}

// Run simulates executing a command.
func (stub *SSHServiceStub) Run(ctx context.Context, cmd string, options ...sshexec.SessionOption) error {
	// Simulate work and context cancellation
	select {
	case <-time.After(300 * time.Millisecond): // Simulate some work
		if stub.RunFunc != nil {
			return stub.RunFunc(ctx, cmd, options...)
		}
		return fmt.Errorf("Run not implemented")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// ReadFile simulates reading a file from the in-memory file system.
func (stub *SSHServiceStub) ReadFile(path string) ([]byte, error) {
	if stub.WriteFileFunc != nil {
		return stub.ReadFileFunc(path)
	}
	return nil, fmt.Errorf("ReadFile not implemented")
}

// WriteFile simulates writing a file to the in-memory file system.
func (stub *SSHServiceStub) WriteFile(path string, data []byte) error {
	if stub.WriteFileFunc != nil {
		return stub.WriteFileFunc(path, data)
	}
	return fmt.Errorf("WriteFile not implemented")
}

// Host returns the configured hostname.
func (stub *SSHServiceStub) Host() string {
	return stub.hostName
}
