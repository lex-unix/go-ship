package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/docker"
	"neite.dev/go-ship/internal/ssh"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup your servers by installing Docker and Caddy",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("setting up your servers...")

		err := docker.IsInstalled().Run()
		if err != nil {
			fmt.Println("Error running `docker --version` locally. Make sure you have docker installed.")
			return
		}

		err = docker.IsRunning().Run()
		if err != nil {
			fmt.Println("Error running `docker version` locally. Make sure docker daemon is running.")
			return
		}

		fmt.Println("reading your config file...")

		if !config.IsExists() {
			fmt.Printf("Could not find your config file. Make sure to run `goship init` first.")
			return

		}

		userCfg, err := config.ReadConfig()
		if err != nil {
			fmt.Println("Could not read your config file")
		}

		commitHash, err := latestCommitHash()
		if err != nil {
			fmt.Println(err)
			return
		}
		imgName := fmt.Sprintf("%s:%s", userCfg.Registry.Image, commitHash)
		fmt.Println(imgName)

		err = docker.BuildImage(imgName, "").Run()
		if err != nil {
			fmt.Println("Error running `docker build`. Could not build your image.")
			return
		}

		err = docker.PushToHub(imgName).Run()
		if err != nil {
			fmt.Println("error running `docker push`. Could not push tag to docker hub.")
			return
		}

		// setup connection with server
		client, err := ssh.NewConnection(userCfg)
		if err != nil {
			fmt.Println("error establishing connection with server.")
			return
		}
		defer client.Close()

		fmt.Println("Connected to server")

		err = docker.IsInstalled().RunSSH(client)
		if err != nil {
			switch {
			// command `docker` could not be found meaning the docker is not installed
			case errors.Is(err, ssh.ErrExit):
				fmt.Println("docker is not installed; installing...")
				err := installDocker(client)
				if err != nil {
					fmt.Println("could not install docker on your server")
					return
				}
			default:
				fmt.Println("could not check if docker is intsalled on the server")
				return
			}
		}

		err = docker.PullFromHub(imgName).RunSSH(client)
		if err != nil {
			fmt.Println("could not pull your image from DockerHub on the server")
			return
		}

		// because it's the setup we can run container instead of starting or restarting it
		err = docker.RunContainer(3000, userCfg.Service, imgName).RunSSH(client)
		if err != nil {
			fmt.Println("could not run your container on the server")
			return
		}

	},
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
