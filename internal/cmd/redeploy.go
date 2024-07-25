package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
)

func init() {
	rootCmd.AddCommand(redeployCmd)
}

var redeployCmd = &cobra.Command{
	Use:   "redeploy",
	Short: "Redeploy your app to the servers",
	Run: func(cmd *cobra.Command, args []string) {
		app, err := app.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		defer app.CloseClients()

		if err := app.PrepareImgForRemote(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := app.RemoveRunningContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := app.LatestRemoteContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
