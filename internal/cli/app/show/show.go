package show

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/lex-unix/faino/internal/cli/cliutil"
)

func NewCmdShow(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show app container on servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := f.App()
			if err != nil {
				return err
			}

			info, err := app.ShowServiceInfo(ctx)
			if err != nil {
				return err
			}

			for host, output := range info {
				fmt.Printf("Host %s:\n%s\n", host, output)
			}

			return nil
		},
	}

	return cmd
}
