package init

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/template"
)

func NewCmdInit(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create shipit.yaml file in current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := template.TemplateFS.ReadFile("templates/shipit.yaml")
			if err != nil {
				return err
			}
			err = os.WriteFile("shipit.yaml", data, 0644)
			if err != nil {
				return err
			}
			logging.Info("created shipit.yaml in current directory")
			return nil
		},
	}

	cliutil.DisableConfigLoading(cmd)

	return cmd
}
