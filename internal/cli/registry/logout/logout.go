package logout

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/lex-unix/faino/internal/cli/cliutil"
)

func NewCmdLogout(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			err = app.StopProxy(ctx)
			if err := app.RegistryLogout(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
