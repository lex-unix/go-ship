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
	rootCmd.AddCommand(stopCmd)
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop app on servers",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.ReadConfig()
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to read your config\n")
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

		err = docker.StopContainer(cfg.Service).RunSSH(sshClient)
		if err != nil {
			fmt.Fprint(os.Stderr, "Unable to stop container")
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
