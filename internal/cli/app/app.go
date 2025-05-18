package app

import (
	"context"

	"github.com/spf13/cobra"
	execCmd "neite.dev/go-ship/internal/cli/app/exec"
	restartCmd "neite.dev/go-ship/internal/cli/app/restart"
	showCmd "neite.dev/go-ship/internal/cli/app/show"
	startCmd "neite.dev/go-ship/internal/cli/app/start"
	stopCmd "neite.dev/go-ship/internal/cli/app/stop"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdApp(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage application on servers",
	}

	cmd.AddCommand(showCmd.NewCmdShow(ctx, f))
	cmd.AddCommand(stopCmd.NewCmdStop(ctx, f))
	cmd.AddCommand(startCmd.NewCmdStart(ctx, f))
	cmd.AddCommand(restartCmd.NewCmdRestart(ctx, f))
	cmd.AddCommand(execCmd.NewCmdExec(ctx, f))

	return cmd
}
