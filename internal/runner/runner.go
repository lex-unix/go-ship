package runner

import (
	"errors"
	"fmt"

	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/docker"
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

	sshClients := lazyloader.Load(func() ([]*ssh.Client, error) {
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
	for _, client := range g.sshClients.Get() {
		client.Close()
	}
}

func (r *runner) PrepareImgForRemote() error {
	appVersion, err := latestCommitHash()
	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	if err != nil {
		return fmt.Errorf("Unable to get most recent commit hash: %s", err)
	}

	err = docker.BuildImage(imgWithTag, r.config.Dockerfile).Run()
	if err != nil {
		return fmt.Errorf("Unable to build image: %s", err)
	}

	err = docker.Tag(imgWithTag, registryImg).Run()
	if err != nil {
		return fmt.Errorf("Unable to tag image: %s", err)
	}

	err = docker.PushToHub(registryImg).Run()
	if err != nil {
		return fmt.Errorf("Unable to push built image to registry: %s", err)
	}

	return nil
}

func (r *runner) RunRemoteContainer() error {
	appVersion, err := latestCommitHash()
	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	err = docker.PullFromHub(registryImg).RunSSH(r.sshClients.Get()[0])
	if err != nil {
		return fmt.Errorf("Unable to pull image from the registry: %s", err)

	}

	err = docker.RunContainer(3000, r.config.Service, registryImg).RunSSH(r.sshClients.Get()[0])
	if err != nil {
		return fmt.Errorf("Could not run your container on the server: %s", err)
	}

	return nil
}

func (r *runner) IntstallDocker() error {
	err := docker.IsInstalled().RunSSH(r.sshClients.Get()[0])
	if err != nil {
		switch {
		// command `docker` could not be found meaning the docker is not installed
		case errors.Is(err, ssh.ErrExit):
			// fmt.Println("docker is not installed; installing...")
			err := installDocker(r.sshClients.Get()[0])
			if err != nil {
				return fmt.Errorf("Unable install Docker on your server: %s", err)
			}
		default:

			return fmt.Errorf("Unable not check if Docker is intsalled on the server: %s", err)
		}
	}
	return nil
}

func installDocker(client *ssh.Client) error {
	sftpClient, err := client.NewSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	err = sftpClient.TransferExecutable("./scripts/setup.sh", "setup.sh")
	if err != nil {
		return err
	}

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run("./setup.sh"); err != nil {
		return err
	}
	return nil
}
