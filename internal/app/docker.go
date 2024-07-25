package app

import (
	"errors"
	"fmt"

	"neite.dev/go-ship/internal/commands"
	"neite.dev/go-ship/internal/ssh"
)

func (a *app) PrepareImgForRemote() error {
	appVersion, err := commands.CommitHash()
	if err != nil {
		return fmt.Errorf("unable to get most recent commit hash: %w", err)
	}

	imgWithTag := fmt.Sprintf("%s:%s", a.config.Image, appVersion)
	registryImg := fmt.Sprintf("%s/%s", a.config.Registry.Server, imgWithTag)

	r, err := a.newLocalRunner()
	if err != nil {
		return err
	}

	if err := r.local(commands.Docker("build -t", imgWithTag, a.config.Dockerfile)); err != nil {
		return fmt.Errorf("unable to build image: %w", err)
	}

	if err := r.local(commands.Docker("tag", imgWithTag, registryImg)); err != nil {
		return fmt.Errorf("unable to tag image: %w", err)
	}

	if err := r.local(commands.Docker("push", registryImg)); err != nil {
		return fmt.Errorf("unable to push built image to registry: %w", err)
	}

	return nil
}

func (a *app) LatestRemoteContainer() error {
	appVersion, err := commands.CommitHash()
	if err != nil {
		return fmt.Errorf("unable to get most recent commit hash: %w", err)
	}
	return a.RunRemoteContainer(appVersion)
}

func (a *app) RunRemoteContainer(version string) error {
	imgWithTag := fmt.Sprintf("%s:%s", a.config.Image, version)
	registryImg := fmt.Sprintf("%s/%s", a.config.Registry.Server, imgWithTag)

	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}

	if err := r.overSSH(commands.Docker("pull", registryImg)); err != nil {
		return fmt.Errorf("unable to pull image from the registry: %w", err)
	}

	portMap := fmt.Sprintf("%d:%d", 3000, 3000)
	if err := r.overSSH(commands.Docker("run", "-d", "-p", portMap, "--name", a.config.Service, registryImg)); err != nil {
		return fmt.Errorf("could not run your container on the server: %w", err)
	}

	return nil
}

func (a *app) RemoveRunningContainer() error {
	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}

	if err := r.overSSH(commands.Docker("stop", a.config.Service)); err != nil {
		return fmt.Errorf("unable to stop your container on the server: %w", err)
	}

	if err := r.overSSH(commands.Docker("rm", a.config.Service)); err != nil {
		return fmt.Errorf("unable to delete your container on the server: %w", err)
	}

	return nil
}

func (a *app) InstallDocker() error {
	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}

	err = r.overSSH(commands.IsDockerInstalled())
	if err != nil {
		if errors.Is(err, ssh.ErrExit) {
			return installDocker(a.sshClients.Load()[0])
		}
		return fmt.Errorf("unable to check if Docker is installed on the server: %w", err)
	}
	return nil
}

func (a *app) ShowContainers() error {
	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}
	return r.overSSH(commands.Docker("ps"))
}

func (a *app) StopContainer() error {
	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}
	return r.overSSH(commands.Docker("stop", a.config.Service))
}

func (a *app) StartContainer() error {
	r, err := a.newSSHRunner()
	if err != nil {
		return err
	}
	return r.overSSH(commands.Docker("start", a.config.Service))
}

func (a *app) newLocalRunner() (*runner, error) {
	return initRunner(withStdout(&a.stdout), withStderr(&a.stderr))
}

func (a *app) newSSHRunner() (*runner, error) {
	clients := a.sshClients.Load()
	if err := a.sshClients.Error(); err != nil {
		return nil, fmt.Errorf("could not establish connection to the server: %w", err)
	}
	return initRunner(withClient(clients[0]), withStdout(&a.stdout), withStderr(&a.stderr))
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
