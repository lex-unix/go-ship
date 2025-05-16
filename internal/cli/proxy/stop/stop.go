package stop

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdStop(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.App.StopProxy(ctx); err != nil {
				return nil
			}
			logging.Info("proxy container stopped on servers")
			return nil
		},
	}

	return cmd
}
