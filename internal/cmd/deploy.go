package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/docker"
	"neite.dev/go-ship/internal/ssh"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your app to the servers",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ReadConfig()
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to read your config\n")
			fmt.Fprint(os.Stderr, err)
			return
		}

		appVersion, err := latestCommitHash()
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to get most recent commit hash\n")
			fmt.Fprint(os.Stderr, err)
			return
		}

		imgWithTag := fmt.Sprintf("%s:%s", cfg.Image, appVersion)

		err = docker.BuildImage(imgWithTag, cfg.Dockerfile).Run()
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to build the image")
			fmt.Fprint(os.Stderr, err)
			return
		}

		registryImg := fmt.Sprintf("%s/%s", cfg.Registry.Server, imgWithTag)

		err = docker.Tag(imgWithTag, registryImg).Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to tag image %s\n", imgWithTag)
			fmt.Fprint(os.Stderr, err)
			return
		}

		err = docker.PushToHub(registryImg).Run()
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to push built image to registry\n")
			fmt.Fprint(os.Stderr, err)
			return
		}

		sshClient, err := ssh.NewConnection(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to establish connection with remote server: %s", cfg.Servers[0])
			fmt.Fprint(os.Stderr, err)
			return
		}
		defer sshClient.Close()

		err = docker.PullFromHub(registryImg).RunSSH(sshClient)
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to pull image from the registry")
			fmt.Fprint(os.Stderr, err)
			return
		}

		err = docker.RunContainer(3000, cfg.Service, registryImg).RunSSH(sshClient)
		if err != nil {
			fmt.Fprint(os.Stderr, "Could not run your container")
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
