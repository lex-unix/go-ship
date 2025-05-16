package stop

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdStop(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			txman, err := f.Txman()
			if err != nil {
				return err
			}

			app := f.App(app.WithTxManager(txman))
			err = app.StopProxy(ctx)
			if err := app.StopProxy(ctx); err != nil {
				return nil
			}
			logging.Info("proxy container stopped on servers")
			return nil
		},
	}

	return cmd
}
