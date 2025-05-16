package init

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func NewCmdInit(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create shipit.yaml file in current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := f.App()
			if err := app.CreateConfig(); err != nil {
				return err
			}
			logging.Info("created shipit.yaml in current directory")
			return nil
		},
	}

	return cmd
}
