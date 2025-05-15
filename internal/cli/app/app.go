package app

import (
	"context"

	"github.com/spf13/cobra"
	showCmd "neite.dev/go-ship/internal/cli/app/show"
	startCmd "neite.dev/go-ship/internal/cli/app/start"
	stopCmd "neite.dev/go-ship/internal/cli/app/stop"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdApp(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage application on the servers",
	}

	cmd.AddCommand(showCmd.NewCmdShow(ctx, f))
	cmd.AddCommand(stopCmd.NewCmdStop(ctx, f))
	cmd.AddCommand(startCmd.NewCmdStart(ctx, f))

	return cmd
}
