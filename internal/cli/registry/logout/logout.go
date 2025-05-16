package logout

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdLogout(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			txman, err := f.Txman()
			if err != nil {
				return err
			}

			app := f.App(app.WithTxManager(txman))
			err = app.StopProxy(ctx)
			if err := app.RegistryLogout(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
