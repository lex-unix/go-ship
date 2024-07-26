package app

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/lazyloader"
	"neite.dev/go-ship/internal/ssh"
)

type app struct {
	config     *config.UserConfig
	sshClients *lazyloader.Loader[[]*ssh.Client]
	stderr     bytes.Buffer
	stdout     bytes.Buffer
}

func New() (*app, error) {
	config, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to read `goship.yaml` file: %s", err)
	}

	sshClients := newLazySSHClientPool(config)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := app{
		config:     config,
		sshClients: sshClients,
		stdout:     stdout,
		stderr:     stderr,
	}

	return &app, nil
}

func (a *app) Stdout() string {
	return a.stdout.String()
}

func (a *app) Stderr() string {
	return a.stderr.String()
}

func (a *app) CloseClients() {
	for _, client := range a.sshClients.Load() {
		client.Close()
	}
}

func (a *app) runLocal(c string) error {
	cmd := exec.Command("sh", "-c", c)
	cmd.Stderr = &a.stdout
	cmd.Stderr = &a.stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (a *app) runOverSSH(c string) error {
	var wg sync.WaitGroup

	for _, client := range a.sshClients.Load() {
		wg.Add(1)
		go runOverSSH(&wg, c, client)

	}
	wg.Wait()
	return nil
}

func runOverSSH(wg *sync.WaitGroup, c string, client *ssh.Client) {
	defer wg.Done()

	sess, err := client.NewSession(os.Stdout, os.Stderr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer sess.Close()

	if err := sess.Run(c); err != nil {
		fmt.Println(err)
	}
}

func newLazySSHClientPool(config *config.UserConfig) *lazyloader.Loader[[]*ssh.Client] {
	return lazyloader.New(func() ([]*ssh.Client, error) {
		sshClients := make([]*ssh.Client, 0, len(config.Servers))
		for _, server := range config.Servers {
			client, err := ssh.NewConnection(server, config.SSH.Port)
			if err != nil {
				return nil, fmt.Errorf("unable to establish connection to %s: %s", server, err)
			}
			sshClients = append(sshClients, client)
		}
		return sshClients, nil
	})
}
