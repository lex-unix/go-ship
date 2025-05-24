package sshexec

import "fmt"

type pipeError struct {
	fd  fd
	err error
}

func (e pipeError) Error() string {
	var d string
	switch e.fd {
	case fdStdin:
		d = "stdin"
	case fdStdout:
		d = "stdout"
	case fdStderr:
		d = "stderr"
	}
	return fmt.Sprintf("%s pipe: %s", d, e.err)
}

type CmdNotFoundErr struct {
	err     error
	command string
	args    string
}

func (e CmdNotFoundErr) Error() string {
	return fmt.Sprintf("command %q not found", e.command)
}

func (e CmdNotFoundErr) Unwrap() error {
	return e.err
}
