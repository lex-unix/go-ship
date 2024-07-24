package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

func init() {
	rootCmd.AddCommand(redeployCmd)
}

var redeployCmd = &cobra.Command{
	Use:   "redeploy",
	Short: "Redeploy your app to the servers",
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

		if err := r.RemoveRunningContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.LatestRemoteContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
