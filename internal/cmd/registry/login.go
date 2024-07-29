package registry

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login into registry",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.RegistryLogin(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
