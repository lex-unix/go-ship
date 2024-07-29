package runner

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

type runner struct {
	config     *config.UserConfig
	sshClients *lazyloader.Loader[[]*ssh.Client]
	stderr     bytes.Buffer
	stdout     bytes.Buffer
}

func New() (*runner, error) {
	config, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to read `goship.yaml` file: %s", err)
	}

	sshClients := newLazySSHClientPool(config)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := runner{
		config:     config,
		sshClients: sshClients,
		stdout:     stdout,
		stderr:     stderr,
	}

	return &app, nil
}

func (r *runner) Stdout() string {
	return r.stdout.String()
}

func (r *runner) Stderr() string {
	return r.stderr.String()
}

func (r *runner) CloseClients() {
	for _, client := range r.sshClients.Load() {
		client.Close()
	}
}

func (r *runner) runLocal(c string) error {
	cmd := exec.Command("sh", "-c", c)
	cmd.Stderr = &r.stdout
	cmd.Stderr = &r.stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (r *runner) runOverSSH(c string) error {
	var wg sync.WaitGroup

	for _, client := range r.sshClients.Load() {
		fmt.Printf("Host: %s\n\n", client.Address())
		runOverSSH(&wg, c, client)
		fmt.Printf("\n\n\n")

	}
	// wg.Wait()
	return nil
}

func runOverSSH(wg *sync.WaitGroup, c string, client *ssh.Client) {
	// defer wg.Done()

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
