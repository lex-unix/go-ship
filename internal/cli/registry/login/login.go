package login

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/lex-unix/faino/internal/cli/cliutil"
)

func NewCmdLogin(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to registry",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			err = app.StopProxy(ctx)
			if err := app.RegistryLogin(ctx); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}
