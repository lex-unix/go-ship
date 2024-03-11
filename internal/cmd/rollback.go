package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/lockfile"
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
	rootCmd.LocalFlags().String("version", "", "commit hash for rollback")
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "rollback to your app's desired version",
	Run: func(cmd *cobra.Command, args []string) {
		// setup cobra flag
		commitHash, err := cmd.Flags().GetString("version")
		if err != nil {
			fmt.Println("failed to get version flag")
			log.Println(err)
			return
		}

		// pass cobra flag
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

		if !exists {
			fmt.Println("version doesn't exist. Make sure to use a valid commit hash")
			return
		}

		userCfg, err := config.ReadConfig()
		if err != nil {
			fmt.Println("Could not read your config file")
		}

		imgName := fmt.Sprintf("%s:%s", userCfg.Registry.Image, commitHash)

		// pull from hub with version tag
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
	},
}
