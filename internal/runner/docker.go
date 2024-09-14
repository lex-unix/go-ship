package runner

import (
	"errors"
	"fmt"
	"strings"

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

	labels := []string{"--label traefik.enable=true", "--label traefik.http.routers.myapp.entrypoints=web", "--label traefik.http.routers.myapp.rule='PathPrefix(`/`)'"}

	if err := r.runOverSSH(commands.Docker("run", "-d", strings.Join(labels, " "), "--name", r.config.Service, registryImg)); err != nil {
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
			// TODO: install docker via script
			fmt.Println(err)
			return nil
		}
		return fmt.Errorf("unable to check if Docker is installed on the server: %w", err)
	}
	return nil
}

func (r *runner) ShowAppInfo() error {
	results, err := r.RunOverSSH(commands.Docker("ps", "--filter", "name="+r.config.Service))
	if err != nil {
		return err
	}

	for result := range results {
		if result.Err != nil {
			fmt.Printf("Error on host: %s\n", result.Err)
			continue
		}
		fmt.Printf("Host: %s\n%s\n\n", result.Host, result.Stdout)
	}

	return nil
}

func (r *runner) StopContainer() error {
	results, err := r.RunOverSSH(commands.Docker("stop", r.config.Service))
	if err != nil {
		return err
	}

	for result := range results {
		if result.Err != nil {
			fmt.Printf("Failed to stop application on %s: %s", result.Host, result.Err)
			continue
		}
		fmt.Printf("Stopped application on %s\n", result.Host)
	}
	return nil
}

func (r *runner) StartContainer() error {
	err := r.runOverSSH(commands.Docker("start", r.config.Service))
	if err != nil {
		return err
	}
	return nil
}

func (r *runner) Deploy() error {
	if err := r.PrepareImgForRemote(); err != nil {
		return err
	}

	if err := r.RegistryLogin(); err != nil {
		return err
	}

	if err := r.RunTraefik(); err != nil {
		return err
	}

	if err := r.LatestRemoteContainer(); err != nil {
		return err
	}

	r.CloseClients()

	return nil
}

func (r *runner) Setup() error {
	if err := r.InstallDocker(); err != nil {
		return err
	}

	if err := r.PrepareImgForRemote(); err != nil {
		return err
	}

	if err := r.RegistryLogin(); err != nil {
		return err
	}

	if err := r.RunTraefik(); err != nil {
		return err
	}

	if err := r.LatestRemoteContainer(); err != nil {
		return err
	}

	return nil
}
