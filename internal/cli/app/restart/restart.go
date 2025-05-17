package restart

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdRestart(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart app container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.RestartService(ctx); err != nil {
				return err
			}
			logging.Info("app container restarted on servers")
			return nil
		},
	}

	return cmd
}
