package rollback

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdRollback(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to your app's desired version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("the version of the app to rollback to is required")
			}
			version := args[0]

			txman, err := f.Txman()
			if err != nil {
				return err
			}

			app := f.App(app.WithTxManager(txman))
			if err := app.Rollback(ctx, version); err != nil {
				return err
			}
			logging.Infof("app rolled back to version %s", version)
			return nil
		},
	}

	return cmd
}
