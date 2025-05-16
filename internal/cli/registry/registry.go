package registry

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	loginCmd "neite.dev/go-ship/internal/cli/registry/login"
	logoutCmd "neite.dev/go-ship/internal/cli/registry/logout"
)

func NewCmdRegistry(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "registry",
		Short: "Manage registry",
	}

	cmd.AddCommand(loginCmd.NewCmdLogin(ctx, f))
	cmd.AddCommand(logoutCmd.NewCmdLogout(ctx, f))

	return cmd
}
