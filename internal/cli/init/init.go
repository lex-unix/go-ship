package init

import (
	"context"
	"os"

	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
	"github.com/lex-unix/faino/internal/template"
	"github.com/spf13/cobra"
)

func NewCmdInit(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create faino.yaml file in current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := template.TemplateFS.ReadFile("templates/faino.yaml")
			if err != nil {
				return err
			}
			err = os.WriteFile("faino.yaml", data, 0644)
			if err != nil {
				return err
			}
			logging.Info("created faino.yaml in current directory")
			return nil
		},
	}

	cliutil.DisableConfigLoading(cmd)

	return cmd
}
