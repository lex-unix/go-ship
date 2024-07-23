package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

func init() {
	rootCmd.AddCommand(deployCmd)
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your app to the servers",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		defer r.CloseClients()

		if err := r.PrepareImgForRemote(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.RunRemoteContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
