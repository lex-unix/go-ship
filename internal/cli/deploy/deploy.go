package deploy

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdDeploy(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy your app to the servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			if err := app.Deploy(ctx); err != nil {
				return err
			}
			logging.Info("app deployed to servers")
			return nil
		},
	}

	return cmd
}
