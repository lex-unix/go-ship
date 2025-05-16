package logout

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdLogout(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.App.RegistryLogout(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
