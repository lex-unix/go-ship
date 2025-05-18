package sshexec

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"neite.dev/go-ship/internal/logging"
)

type fd uint8

const (
	fdStdin fd = iota
	fdStdout
	fdStderr
)

var privateKeys = []string{
	"id_rsa",
	"id_ecdsa",
	"id_ecdsa_sk",
	"id_ed25519",
	"id_ed25519_sk",
}

type Service interface {
	Run(ctx context.Context, cmd string, options ...RunOption) error

	ReadFile(path string) ([]byte, error)

	WriteFile(path string, data []byte) error

	Host() string
}

type SSH struct {
	client *ssh.Client
	host   string
}

func New(host, user string, port int64) (*SSH, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	sshDir := filepath.Join(homeDir, ".ssh")
	hostkeyCallback, err := knownhosts.New(filepath.Join(sshDir, "known_hosts"))
	if err != nil {
		return nil, err
	}

	var authMethod ssh.AuthMethod
	var socketErr error
	// try to use ssh agent for authentication like 1password
	socketPath := os.Getenv("SSH_AUTH_SOCK")
	if socketPath != "" {
		var socket net.Conn
		socket, socketErr = net.Dial("unix", socketPath)
		if socketErr == nil {
			sshAgent := agent.NewClient(socket)
			authMethod = ssh.PublicKeysCallback(sshAgent.Signers)
		}
	}

	// if there is an error with ssh agent, will try to use private keys
	if socketErr != nil {
		var signers []ssh.Signer
		for _, pkeyFile := range privateKeys {
			if signer, err := parsePrivateKey(filepath.Join(sshDir, pkeyFile)); err == nil {
				signers = append(signers, signer)
			}
		}
		if len(signers) == 0 {
			return nil, fmt.Errorf("ssh: no auth method detected")
		}
		authMethod = ssh.PublicKeys(signers...)
	}

	config := &ssh.ClientConfig{
		User:            user,
		Auth:            []ssh.AuthMethod{authMethod},
		HostKeyCallback: hostkeyCallback,
	}

	client, err := ssh.Dial("tcp", formatAddress(host, port), config)
	if err != nil {
		return nil, err
	}

	return &SSH{client: client, host: host}, nil
}

func (s *SSH) Host() string {
	return s.host
}

func (s *SSH) Run(ctx context.Context, cmd string, options ...RunOption) error {
	opts := RunConfig{
		interactive: false,
		stdout:      &logWriter{host: s.host},
		stderr:      &logWriter{host: s.host},
	}

	for _, opt := range options {
		opt(&opts)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	doneCh := make(chan struct{}, 1)
	// close channel when exiting to signal goroutine below to stop
	defer func() { close(doneCh) }()
	go func() {
		select {
		case <-ctx.Done():
			if err := session.Signal(ssh.SIGTERM); err != nil {
				logging.ErrorHostf(s.host, "failed to stop command: %s", err)
			}
			return
		case <-doneCh:
			return
		}
	}()

	if opts.interactive {
		return s.runInteractiveSession(session, cmd, opts)
	}
	return s.run(session, cmd, opts)
}

func (s *SSH) WriteFile(path string, data []byte) error {
	r := bytes.NewReader(data)
	cmd := fmt.Sprintf("cat > %s", path)
	return s.Run(context.Background(), cmd, WithStdin(r))
}

func (s *SSH) ReadFile(path string) ([]byte, error) {
	var contents bytes.Buffer
	cmd := fmt.Sprintf("cat %s", path)
	// pass noop context to finish reading file
	err := s.Run(context.Background(), cmd, WithStdout(&contents))
	if err != nil {
		var pipeErr *pipeError
		// if error is not copying stdout, return read contents
		if errors.As(err, &pipeErr); pipeErr.fd != fdStdout {
			return contents.Bytes(), nil
		}
		return nil, err
	}

	return contents.Bytes(), nil
}

type RunOption func(c *RunConfig)

type RunConfig struct {
	stdout      io.Writer
	stderr      io.Writer
	stdin       io.Reader
	interactive bool
}

func WithStdout(w io.Writer) RunOption {
	return func(c *RunConfig) {
		c.stdout = w
	}
}

func WithStderr(w io.Writer) RunOption {
	return func(c *RunConfig) {
		c.stderr = w
	}
}

func WithStdin(r io.Reader) RunOption {
	return func(c *RunConfig) {
		c.stdin = r
	}
}

func WithPty() RunOption {
	return func(c *RunConfig) {
		c.interactive = true
		c.stdin = os.Stdin
		c.stdout = os.Stdout
		c.stderr = os.Stderr
	}
}

func formatAddress(host string, port int64) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func parsePrivateKey(path string) (ssh.Signer, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	pemBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ssh.ParsePrivateKey(pemBytes)
}
