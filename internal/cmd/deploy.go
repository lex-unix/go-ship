package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/docker"
	"neite.dev/go-ship/internal/lockfile"
	"neite.dev/go-ship/internal/ssh"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your latest image from Docker Hub",
	Run: func(cmd *cobra.Command, args []string) {
		// check if version lock file exists
		lfPath, err := lockfile.LockPath()
		if err != nil {
			fmt.Println(err)
			return
		}

		if _, err := os.Stat(lfPath); err != nil {
			fmt.Println("Missing lock file, please do `goship setup`.")
			return
		}

		// take the latest commitHash
		commitHash, err := latestCommitHash()
		if err != nil {
			fmt.Printf("Error running `git rev-parse --short HEAD`. Error: %q", err)
			return
		}

		file, err := lockfile.OpenFile()
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()

		exists, err := lockfile.VersionExists(file, commitHash)
		if err != nil {
			fmt.Printf("could not check version in lock file. Error: %s", err)
			return
		}

		if exists {
			fmt.Println("already on the most recent version.")
			return
		}

		userCfg, err := config.ReadConfig()
		if err != nil {
			fmt.Println("Could not read your config file")
		}

		imgName := fmt.Sprintf("%s/%s:%s", userCfg.Registry.Username, userCfg.Image, commitHash)

		err = docker.BuildImage(imgName, userCfg.Dockerfile).Run()
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

		err = docker.PullFromHub(imgName).RunSSH(client)
		if err != nil {
			fmt.Println("could not pull your image from DockerHub on the server")
			return
		}

		container := userCfg.Service

		err = docker.StopContainer(container).RunSSH(client)
		if err != nil {
			fmt.Println("could not stop your container on the server")
			return
		}

		err = docker.RemoveContainer(container).RunSSH(client)
		if err != nil {
			fmt.Println("could not delete your container on the server")
			return
		}

		// because it's the setup we can run container instead of starting or restarting it
		err = docker.RunContainer(3000, container, imgName).RunSSH(client)
		if err != nil {
			fmt.Println("could not run your container on the server")
			return
		}

		f, err := os.OpenFile(lfPath, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Could not open `goship-lock.json`")
			return
		}

		defer f.Close()

		commitMsg, err := latestCommitMsg()
		if err != nil {
			fmt.Println(err)
			return
		}

		data := lockfile.LockVersion{
			Version: commitHash,
			Image:   imgName,
			Message: commitMsg,
			Date:    now(),
		}

		if err := lockfile.Write(f, data); err != nil {
			fmt.Printf("could not write to lockfile\n. Error: %s", err)
			return

		}

	},
}
