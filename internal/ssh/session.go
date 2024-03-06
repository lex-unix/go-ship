package ssh

import (
	"errors"
	"io"

	"golang.org/x/crypto/ssh"
)

var (
	ErrExit = errors.New("command exit error")
	//other errors; prefix them with Err
)

type sshOption func(options *sessionOptions) error

type sessionOptions struct {
	stdout io.Writer
	stderr io.Writer
}

type Session struct {
	sshSess *ssh.Session
}

func WithStdout(stdout io.Writer) sshOption {
	return func(options *sessionOptions) error {
		if stdout == nil {
			return errors.New("stdout pipe could not be set to nil")
		}
		options.stdout = stdout
		return nil
	}
}

func WithStderr(stderr io.Writer) sshOption {
	return func(options *sessionOptions) error {
		if stderr == nil {
			return errors.New("stderr pipe could not be set to nil")
		}
		options.stdout = stderr
		return nil
	}
}

func (s Session) Run(cmd string) error {
	return s.sshSess.Run(cmd)

}

func (s Session) Close() error {
	return s.sshSess.Close()
}
