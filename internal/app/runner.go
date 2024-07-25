package app

import (
	"fmt"
	"io"
	"os/exec"

	"neite.dev/go-ship/internal/ssh"
)

type options struct {
	overSSH   bool
	out       io.Writer
	stdout    io.Writer
	stderr    io.Writer
	sshClient *ssh.Client
}

type option func(opts *options) error

func withOut(out io.Writer) option {
	return func(opts *options) error {
		opts.out = out
		return nil
	}
}

func withStdout(stdout io.Writer) option {
	return func(opts *options) error {
		opts.stdout = stdout
		return nil
	}
}

func withStderr(stderr io.Writer) option {
	return func(opts *options) error {
		opts.stderr = stderr
		return nil
	}
}

func withSSHClient(client *ssh.Client) option {
	return func(opts *options) error {
		opts.overSSH = true
		opts.sshClient = client
		return nil
	}
}

func buildOpts(opts ...option) (*options, error) {
	var options options
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}
	return &options, nil
}

func run(cmd string, opts ...option) error {
	options, _ := buildOpts(opts...)

	if options.overSSH {
		return runOverSSH(cmd, options)
	} else {
		return runLocally(cmd, options)
	}
}

func runLocally(subcmd string, options *options) error {
	cmd := exec.Command("sh", "-c", subcmd)
	if options.out != nil {
		cmd.Stdout = options.out
		cmd.Stderr = options.out
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func runOverSSH(subcmd string, options *options) error {
	client := options.sshClient

	var session *ssh.Session
	var err error
	if options.out != nil {
		session, err = client.NewSession(ssh.WithStderr(options.out), ssh.WithStdout(options.out))
		if err != nil {
			fmt.Println(err)
			return err
		}
	} else {
		session, err = client.NewSession()
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	err = session.Run(subcmd)
	if err != nil {
		return err
	}

	return nil
}
