package app

import (
	"bytes"
	"errors"
	"fmt"

	"neite.dev/go-ship/internal/commands"
	"neite.dev/go-ship/internal/ssh"
)

func (a *app) PrepareImgForRemote() error {
	appVersion, err := latestCommitHash()
	if err != nil {
		return fmt.Errorf("Unable to get most recent commit hash: %s", err)
	}

	imgWithTag := fmt.Sprintf("%s:%s", a.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", a.config.Registry.Server, imgWithTag)

	err = run(commands.Docker("build -t", imgWithTag, a.config.Dockerfile))
	if err != nil {
		return fmt.Errorf("Unable to build image: %s", err)
	}

	err = run(commands.Docker("tag", imgWithTag, registryImg))
	if err != nil {
		return fmt.Errorf("Unable to tag image: %s", err)
	}

	err = run(commands.Docker("push", registryImg))
	if err != nil {
		return fmt.Errorf("Unable to push built image to registry: %s", err)
	}

	return nil
}

func (a *app) LatestRemoteContainer() error {
	appVersion, err := latestCommitHash()
	if err != nil {
		return fmt.Errorf("Unable to get most recent commit hash: %s", err)
	}
	return a.RunRemoteContainer(appVersion)
}

func (a *app) RunRemoteContainer(version string) error {
	imgWithTag := fmt.Sprintf("%s:%s", a.config.Image, version)
	registryImg := fmt.Sprintf("%s/%s", a.config.Registry.Server, imgWithTag)

	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	// TODO: should check for image on the remote first before pulling from registry
	err := run(commands.Docker("pull", registryImg), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to pull image from the registry: %s", err)

	}

	portMap := fmt.Sprintf("%d:%d", 3000, 3000)
	err = run(
		commands.Docker("run", "-d", "-p", portMap, "--name", a.config.Service, registryImg),
		withSSHClient(clients[0]),
	)
	if err != nil {
		return fmt.Errorf("Could not run your container on the server: %s", err)
	}

	return nil
}

func (a *app) RemoveRunningContainer() error {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("stop", a.config.Service), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to stop your container on the server: %s", err)
	}

	err = run(commands.Docker("rm", a.config.Service), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to delete your container on the server: %s", err)
	}

	return nil
}

func (a *app) IntstallDocker() error {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.IsDockerInstalled(), withSSHClient(clients[0]))
	if err != nil {
		switch {
		// command `docker` could not be found meaning the docker is not installed
		case errors.Is(err, ssh.ErrExit):
			// fmt.Println("docker is not installed; installing...")
			err := installDocker(clients[0])
			if err != nil {
				return fmt.Errorf("Unable install Docker on your server: %s", err)
			}
		default:

			return fmt.Errorf("Unable not check if Docker is intsalled on the server: %s", err)
		}
	}
	return nil
}

func (a *app) ShowContainers() (string, error) {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return "", fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	var out bytes.Buffer
	err := run(commands.Docker("ps"), withSSHClient(clients[0]), withOut(&out))
	if err != nil {
		return out.String(), err
	}

	return out.String(), nil
}

func (a *app) StopContainer() error {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("ps"), withSSHClient(clients[0]))
	if err != nil {
		return err
	}

	return nil
}

func (a *app) StartContainer() error {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("start", a.config.Service), withSSHClient(clients[0]))
	if err != nil {
		return err
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
