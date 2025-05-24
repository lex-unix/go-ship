package registry

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/lex-unix/faino/internal/cli/cliutil"
	loginCmd "github.com/lex-unix/faino/internal/cli/registry/login"
	logoutCmd "github.com/lex-unix/faino/internal/cli/registry/logout"
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
