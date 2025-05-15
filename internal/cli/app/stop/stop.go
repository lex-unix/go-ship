package stop

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdStop(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop app container on server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.App.StopContainer(ctx); err != nil {
				return err
			}
			logging.Info("container is stopped on all servers")
			return nil
		},
	}

	return cmd
}
