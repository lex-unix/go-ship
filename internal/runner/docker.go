package runner

import (
	"errors"
	"fmt"

	"neite.dev/go-ship/internal/commands"
	"neite.dev/go-ship/internal/ssh"
)

func (r *runner) PrepareImgForRemote() error {
	appVersion, err := commands.CommitHash()
	if err != nil {
		return fmt.Errorf("unable to get most recent commit hash: %w", err)
	}

	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	if err := r.runLocal(commands.Docker("build -t", imgWithTag, r.config.Dockerfile)); err != nil {
		return fmt.Errorf("unable to build image: %w", err)
	}

	if err := r.runLocal(commands.Docker("tag", imgWithTag, registryImg)); err != nil {
		return fmt.Errorf("unable to tag image: %w", err)
	}

	if err := r.runLocal(commands.Docker("push", registryImg)); err != nil {
		return fmt.Errorf("unable to push built image to registry: %w", err)
	}

	return nil
}

func (r *runner) LatestRemoteContainer() error {
	appVersion, err := commands.CommitHash()
	if err != nil {
		return fmt.Errorf("unable to get most recent commit hash: %w", err)
	}
	return r.RunRemoteContainer(appVersion)
}

func (r *runner) RunRemoteContainer(version string) error {
	imgWithTag := fmt.Sprintf("%s:%s", r.config.Image, version)
	registryImg := fmt.Sprintf("%s/%s", r.config.Registry.Server, imgWithTag)

	if err := r.runOverSSH(commands.Docker("pull", registryImg)); err != nil {
		return fmt.Errorf("unable to pull image from the registry: %w", err)
	}

	portMap := fmt.Sprintf("%d:%d", 3000, 3000)
	if err := r.runOverSSH(commands.Docker("run", "-d", "-p", portMap, "--name", r.config.Service, registryImg)); err != nil {
		return fmt.Errorf("could not run your container on the server: %w", err)
	}

	return nil
}

func (r *runner) RemoveRunningContainer() error {
	if err := r.runOverSSH(commands.Docker("stop", r.config.Service)); err != nil {
		return fmt.Errorf("unable to stop your container on the server: %w", err)
	}

	if err := r.runOverSSH(commands.Docker("rm", r.config.Service)); err != nil {
		return fmt.Errorf("unable to delete your container on the server: %w", err)
	}

	return nil
}

func (r *runner) InstallDocker() error {
	err := r.runOverSSH(commands.IsDockerInstalled())
	if err != nil {
		if errors.Is(err, ssh.ErrExit) {
			return installDocker(r.sshClients.Load()[0])
		}
		return fmt.Errorf("unable to check if Docker is installed on the server: %w", err)
	}
	return nil
}

func (r *runner) ShowContainers() error {
	return r.runOverSSH(commands.Docker("ps"))
}

func (r *runner) StopContainer() error {
	return r.runOverSSH(commands.Docker("stop", r.config.Service))
}

func (r *runner) StartContainer() error {
	return r.runOverSSH(commands.Docker("start", r.config.Service))
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

	session, err := client.NewSession(nil, nil)
	if err != nil {
		return err
	}
	defer session.Close()

	if err := session.Run("./setup.sh"); err != nil {
		return err
	}
	return nil
}
