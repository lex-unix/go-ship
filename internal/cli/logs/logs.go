package logs

import (
	"context"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
)

func NewCmdLogs(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Fetch logs from you containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			var (
				follow bool
				lines  int
				since  string
				err    error
			)
			follow, err = cmd.Flags().GetBool("follow")
			lines, err = cmd.Flags().GetInt("lines")
			since, err = cmd.Flags().GetString("since")
			if err != nil {
				return err
			}

			if err := f.App.Logs(ctx, follow, lines, since); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolP("follow", "f", false, "Follow logs on servers")
	cmd.PersistentFlags().IntP("lines", "n", 100, "Number of lines to show from each server")
	cmd.PersistentFlags().String("since", "", "Show lines since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")

	return cmd
}
