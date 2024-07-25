package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
)

func init() {
	AppCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start app container on servers",
	Run: func(cmd *cobra.Command, args []string) {
		app, err := app.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := app.StopContainer(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return

		}
	},
}
