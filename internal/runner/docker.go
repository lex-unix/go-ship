package runner

import (
	"bytes"
	"errors"
	"fmt"

	"neite.dev/go-ship/internal/commands"
	"neite.dev/go-ship/internal/ssh"
)

func (r *runner) PrepareImgForRemote() error {
	appVersion, err := latestCommitHash()
	if err != nil {
		return fmt.Errorf("Unable to get most recent commit hash: %s", err)
	}

	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	err = run(commands.Docker("build -t", imgWithTag, r.config.Dockerfile))
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

func (r *runner) LatestRemoteContainer() error {
	appVersion, err := latestCommitHash()
	if err != nil {
		return fmt.Errorf("Unable to get most recent commit hash: %s", err)
	}
	return r.RunRemoteContainer(appVersion)
}

func (r *runner) RunRemoteContainer(version string) error {
	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, version)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	// TODO: should check for image on the remote first before pulling from registry
	err := run(commands.Docker("pull", registryImg), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to pull image from the registry: %s", err)

	}

	portMap := fmt.Sprintf("%d:%d", 3000, 3000)
	err = run(
		commands.Docker("run", "-d", "-p", portMap, "--name", r.config.Service, registryImg),
		withSSHClient(clients[0]),
	)
	if err != nil {
		return fmt.Errorf("Could not run your container on the server: %s", err)
	}

	return nil
}

func (r *runner) RemoveRunningContainer() error {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("stop", r.config.Service), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to stop your container on the server: %s", err)
	}

	err = run(commands.Docker("rm", r.config.Service), withSSHClient(clients[0]))
	if err != nil {
		return fmt.Errorf("Unable to delete your container on the server: %s", err)
	}

	return nil
}

func (r *runner) IntstallDocker() error {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
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

func (r *runner) ShowContainers() (string, error) {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return "", fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	var out bytes.Buffer
	err := run(commands.Docker("ps"), withSSHClient(clients[0]), withOut(&out))
	if err != nil {
		return out.String(), err
	}

	return out.String(), nil
}

func (r *runner) StopContainer() error {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("ps"), withSSHClient(clients[0]))
	if err != nil {
		return err
	}

	return nil
}

func (r *runner) StartContainer() error {
	clients := r.sshClients.Load()
	if err := r.sshClients.Error(); err != nil {
		return fmt.Errorf("Could not establish to connection to the server %s", err)
	}

	err := run(commands.Docker("start", r.config.Service), withSSHClient(clients[0]))
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
