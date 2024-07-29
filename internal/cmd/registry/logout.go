package registry

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/runner"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout from registry",
	Run: func(cmd *cobra.Command, args []string) {
		r, err := runner.New()
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}

		if err := r.RegistryLogout(); err != nil {
			fmt.Fprint(os.Stderr, err)
			return
		}
	},
}
