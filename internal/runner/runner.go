package runner

import (
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
}

type Result struct {
	Err    error
	Stdout []byte
	Host   string
}

func New() (*runner, error) {
	config, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to read `goship.yaml` file: %s", err)
	}

	app := runner{
		config:     config,
		sshClients: newLazySSHClientPool(config),
	}

	return &app, nil
}

func (r *runner) CloseClients() {
	for _, client := range r.sshClients.Load() {
		client.Close()
	}
}

func (r *runner) runLocal(c string) error {
	cmd := exec.Command("sh", "-c", c)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func (r *runner) RunOverSSH(cmd string) (<-chan Result, error) {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return nil, err
	}

	totalClients := len(clients)

	results := make(chan Result, totalClients)
	var wg sync.WaitGroup

	fn := func(client *ssh.Client) {
		defer wg.Done()

		out, err := client.Exec(cmd)
		result := Result{Stdout: out, Host: client.Host, Err: err}
		results <- result
	}

	wg.Add(totalClients)

	for _, client := range clients {
		go fn(client)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results, nil
}

func (r *runner) runOverSSH(c string) error {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, client := range clients {
		wg.Add(1)
		go runOverSSH(&wg, c, client)
	}

	wg.Wait()

	return nil
}

func runOverSSH(wg *sync.WaitGroup, c string, client *ssh.Client) {
	defer wg.Done()

	if _, err := client.Exec(c); err != nil {
		// return err
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
