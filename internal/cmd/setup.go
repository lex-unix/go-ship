package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

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

		imgName := userCfg.Registry.Image
		usrName := userCfg.Registry.Username
		repoName := userCfg.Registry.Reponame

		err = docker.BuildImage(imgName, "").Run()
		if err != nil {
			fmt.Println("Error running `docker build`. Could not build your image.")
		}

		err = docker.RunContainer(3000, imgName, imgName).Run()
		if err != nil {
			fmt.Println("Error running `docker run`. Could not run container.")
		}
		err = docker.LoginToHub(usrName, userCfg.Registry.Password).Run()
		if err != nil {
			fmt.Println("error running `docker login`. Could not login to docker hub.")
			return
		}

		err = docker.RenameImage(imgName, usrName, repoName).Run()
		if err != nil {
			fmt.Println("error running `docker tag`. Could not rename image for docker hub.")
		}

		err = docker.PushToHub(usrName, repoName).Run()
		if err != nil {
			fmt.Println("error running `docker push`. Could not push tag to docker hub.")
		}

		// setup connection with server
		client, err := ssh.NewConnection(userCfg)
		if err != nil {
			fmt.Println("error establishing connection with server.")
			return
		}

		fmt.Println("Connected to server")

		var cmds = []docker.DockerCmd{
			docker.IsInstalled(),
			docker.PullFromHub(usrName, repoName),
			docker.ListImages(),
		}
		for i, cmd := range cmds {
			if err := cmd.RunSSH(client); err != nil {
				switch {
				case errors.Is(err, ssh.ErrExit):
					if i == 0 {
						// call docker install func to run shell script
						if err := docker.Install(client); err != nil {
							fmt.Println("error installing docker.")
							return
						}
						continue
					}
					fmt.Printf("error from running docker command: %v", err)
				default:
					fmt.Printf("error from running docker command: %v", err)
				}
			}
		}

		session, err := client.NewSession(ssh.WithStdout(os.Stdout))
		if err != nil {
			fmt.Printf("error from opening a new session: %v\n", err)
		}

		defer func() {
			if err := session.Close(); err != nil && err != io.EOF {
				fmt.Printf("error from closing session: %v\n", err)
			}
		}()

		if err := session.Run("exit"); err != nil {
			fmt.Printf("error running command `exit`: %v\n", err)
		}

	},
}
