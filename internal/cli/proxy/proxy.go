package proxy

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	logsCmd "neite.dev/go-ship/internal/cli/proxy/logs"
	restartCmd "neite.dev/go-ship/internal/cli/proxy/restart"
	startCmd "neite.dev/go-ship/internal/cli/proxy/start"
	stopCmd "neite.dev/go-ship/internal/cli/proxy/stop"
)

func NewCmdProxy(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Manage proxy container on servers",
	}

	cmd.AddCommand(logsCmd.NewCmdLogs(ctx, f))
	cmd.AddCommand(startCmd.NewCmdStart(ctx, f))
	cmd.AddCommand(stopCmd.NewCmdStop(ctx, f))
	cmd.AddCommand(restartCmd.NewCmdRestart(ctx, f))

	return cmd
}
