package show

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdShow(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show app containers on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.App.ShowAppInfo(ctx); err != nil {
				return err
			}
			return nil
		},
	}

	return cmd
}
