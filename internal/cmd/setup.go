package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

func init() {
	rootCmd.AddCommand(setupCmd)
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup your servers by installing Docker",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()

		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		defer r.CloseClients()

		if err := r.Setup(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

	},
}
