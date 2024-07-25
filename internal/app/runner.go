package app

import (
	"io"
	"os/exec"

	"neite.dev/go-ship/internal/ssh"
)

type runneropts struct {
	client *ssh.Client
	stdout io.Writer
	stderr io.Writer
}

type runneropt func(opts *runneropts) error

type runner struct {
	client *ssh.Client
	stdout io.Writer
	stderr io.Writer
}

func withStdout(stdout io.Writer) runneropt {
	return func(opts *runneropts) error {
		opts.stderr = stdout
		return nil
	}
}

func withClient(client *ssh.Client) runneropt {
	return func(opts *runneropts) error {
		opts.client = client
		return nil
	}
}

func withStderr(stderr io.Writer) runneropt {
	return func(opts *runneropts) error {
		opts.stderr = stderr
		return nil
	}
}

func initRunner(opts ...runneropt) (*runner, error) {
	var options runneropts
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	var r runner
	r.client = options.client
	r.stdout = options.stdout
	r.stderr = options.stderr

	return &r, nil
}

func (r *runner) local(arg string) error {
	cmd := exec.Command("sh", "-c", arg)
	cmd.Stderr = r.stdout
	cmd.Stderr = r.stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (r *runner) overSSH(cmd string) error {
	session, err := r.client.NewSession(r.stdout, r.stderr)
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run(cmd); err != nil {
		return err
	}
	return nil
}
