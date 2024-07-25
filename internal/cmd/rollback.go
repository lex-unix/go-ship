package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "rollback to your app's desired version",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Fprint(os.Stderr, "Please, provide the version of the app to rollback\n")
			return
		}

		appVersion := args[0]

		app, err := app.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		defer app.CloseClients()

		if err := app.RemoveRunningContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := app.RunRemoteContainer(appVersion); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
