package start

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdStart(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start existing proxy container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.StartProxy(ctx); err != nil {
				return err
			}

			logging.Info("proxy container started on servers")
			return nil
		},
	}

	return cmd
}
