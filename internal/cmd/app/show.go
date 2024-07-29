package app

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

func init() {
	AppCmd.AddCommand(showCmd)
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show app containers on the servers",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.ShowAppInfo(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
		fmt.Println(r.Stdout())
	},
}
