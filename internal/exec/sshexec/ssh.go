package sshexec

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"neite.dev/go-ship/internal/logging"
)

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
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	hostkeyCallback, err := knownhosts.New(path.Join(homedir, ".ssh", "known_hosts"))
	if err != nil {
		return nil, err
	}

	var authMethod ssh.AuthMethod
	socketPath := os.Getenv("SSH_AUTH_SOCK")
	// try to use ssh agent for authentication like 1password
	if socketPath != "" {
		socket, err := net.Dial("unix", socketPath)
		// if there is an error, will try to use public key file
		if err == nil {
			sshAgent := agent.NewClient(socket)
			authMethod = ssh.PublicKeysCallback(sshAgent.Signers)
		}
	} else {
		key, err := os.ReadFile(path.Join(homedir, ".ssh", "id_rsa"))
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		authMethod = ssh.PublicKeys(signer)
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

func (s *SSH) Run(ctx context.Context, cmd string, options ...RunOption) error {
	runCfg := RunConfig{}
	for _, opt := range options {
		opt(&runCfg)
	}

	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create new ssh session: %w", err)
	}
	defer session.Close()

	doneCh := make(chan struct{}, 1)
	defer func() { close(doneCh) }()

	go func() {
		select {
		case <-ctx.Done():
			logging.DebugHost(s.host, "signaling remote process to exit...")
			err := session.Signal(ssh.SIGTERM)
			if err != nil {
				logging.ErrorHostf(s.host, "failed to stop remote process: %s", err)
			}
		case <-doneCh:
			return
		}
	}()

	var wg sync.WaitGroup
	stderr, _ := session.StderrPipe()
	stdout, _ := session.StdoutPipe()

	wg.Add(2)
	go s.read(&wg, stdout, runCfg.stdout)
	go s.read(&wg, stderr, runCfg.stderr)

	logging.InfoHostf(s.host, "running command %q", cmd)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command %q: %w", err)
	}

	waitErr := session.Wait()
	wg.Wait()
	if err != nil {
		return fmt.Errorf("failed to execute command %s: %w", cmd, waitErr)
	}

	return nil
}

func (s *SSH) WriteFile(path string, data []byte) error {
	session, err := s.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create new ssh session: %w", err)
	}
	defer session.Close()

	var wg sync.WaitGroup
	stderr, _ := session.StderrPipe()
	stdout, _ := session.StdoutPipe()
	stdin, err := session.StdinPipe()
	if err != nil {
		logging.ErrorHostf(s.host, "failed to write to file: %s", err)
		return fmt.Errorf("failed to get stdin: %w", err)
	}

	wg.Add(2)
	go s.read(&wg, stdout, nil)
	go s.read(&wg, stderr, nil)

	logging.InfoHostf(s.host, "writing to file at %s", path)

	cmd := fmt.Sprintf("cat > %s", path)
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command %q: %w", err)
	}

	_, err = stdin.Write(data)
	if err != nil {
		return err
	}
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin pipe: %w", err)
	}

	waitErr := session.Wait()
	wg.Wait()
	if waitErr != nil {
		logging.ErrorHostf(s.host, "failed to write to file: %s", err)
		return fmt.Errorf("failed to write to stdin: %w", err)
	}

	return nil
}

func (s *SSH) ReadFile(path string) ([]byte, error) {
	session, err := s.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create new ssh session: %w", err)
	}
	defer session.Close()

	var data bytes.Buffer
	var wg sync.WaitGroup
	stderr, _ := session.StderrPipe()
	stdout, _ := session.StdoutPipe()

	readErrCh := make(chan error, 1)
	var read = func(in io.Reader, out io.Writer) {
		defer wg.Done()
		scanner := bufio.NewScanner(in)
		for scanner.Scan() {
			line := scanner.Bytes()
			if _, err := out.Write(line); err != nil {
				logging.Errorf("failed to write file output")
				readErrCh <- err
				return
			}
		}
		if err := scanner.Err(); err != nil {
			readErrCh <- fmt.Errorf("scanner error: %w", err)
			return
		}
		readErrCh <- nil
	}

	wg.Add(2)
	go read(stdout, &data)
	go s.read(&wg, stderr, nil)

	logging.InfoHostf(s.host, "reading file at %s", path)

	cmd := fmt.Sprintf("cat %s", path)
	if err := session.Start(cmd); err != nil {
		return nil, fmt.Errorf("failed to start command %q: %w", err)
	}

	waitErr := session.Wait()
	wg.Wait()
	if waitErr != nil {
		return nil, fmt.Errorf("failed to execute command %s: %w", cmd, waitErr)
	}

	err = <-readErrCh
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}

	return data.Bytes(), nil
}

func (s *SSH) Host() string {
	return s.host
}

func (s *SSH) read(wg *sync.WaitGroup, in io.Reader, out io.Writer) {
	defer wg.Done()
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		line := scanner.Bytes()
		if out != nil {
			line = append(line, '\n')
			if _, err := out.Write(line); err != nil {
				logging.ErrorHostf(s.host, "capture remote command failed: %s", err)
			}
		} else {
			logging.DebugHost(s.host, string(line))
		}
	}
}

type RunOption func(c *RunConfig)

type RunConfig struct {
	stdout io.Writer
	stderr io.Writer
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
func formatAddress(host string, port int64) string {
	return fmt.Sprintf("%s:%d", host, port)
}
