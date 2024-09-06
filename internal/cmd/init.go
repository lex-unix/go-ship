package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create config file",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.CreateConfig(); err != nil {
			fmt.Println(os.Stderr, err)
			return
		}

		fmt.Println("Initialized config file `goship.yaml` in the current current directory")
	},
}
