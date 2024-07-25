package runner

import (
	"fmt"

	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/lazyloader"
	"neite.dev/go-ship/internal/ssh"
)

type runner struct {
	config     *config.UserConfig
	sshClients *lazyloader.Loader[[]*ssh.Client]
}

func New() (*runner, error) {
	cfg, err := config.ReadConfig()
	if err != nil {
		return nil, fmt.Errorf("Unable to read `goship.yaml` file: %s", err)
	}

	sshClients := lazyloader.New(func() ([]*ssh.Client, error) {
		sshClients := make([]*ssh.Client, 0, len(cfg.Servers))
		for _, server := range cfg.Servers {
			client, err := ssh.NewConnection(server, cfg.SSH.Port)
			if err != nil {
				return nil, fmt.Errorf("unable to establish connection to %s: %s", server, err)
			}
			sshClients = append(sshClients, client)

		}
		return sshClients, nil
	})

	return &runner{config: cfg, sshClients: sshClients}, nil
}

func (g *runner) CloseClients() {
	for _, client := range g.sshClients.Load() {
		client.Close()
	}
}
