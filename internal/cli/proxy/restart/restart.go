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
		Short: "Restart proxy container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := f.App.StopProxy(ctx)
			if err != nil {
				logging.Errorf("failed to stop proxy container: %s", err)
				return err
			}

			err = f.App.StartProxy(ctx)
			if err != nil {
				logging.Errorf("failed to start proxy container: %s", err)
				return err
			}

			logging.Info("proxy container restarted")

			return nil
		},
	}

	return cmd
}
