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
	return fmt.Sprintf("%s pipe: %w", d, e.err)
}
