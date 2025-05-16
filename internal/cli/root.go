package cli

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	appCmd "neite.dev/go-ship/internal/cli/app"
	"neite.dev/go-ship/internal/cli/cliutil"
	deployCmd "neite.dev/go-ship/internal/cli/deploy"
	historyCmd "neite.dev/go-ship/internal/cli/history"
	initCmd "neite.dev/go-ship/internal/cli/init"
	logsCmd "neite.dev/go-ship/internal/cli/logs"
	proxyCmd "neite.dev/go-ship/internal/cli/proxy"
	registryCmd "neite.dev/go-ship/internal/cli/registry"
	rollbackCmd "neite.dev/go-ship/internal/cli/rollback"
	"neite.dev/go-ship/internal/logging"
)

func NewRootCmd(ctx context.Context, f *cliutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "shipit",
		Long: "shipit",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("debug") {
				logging.SetDefault(logging.New(os.Stderr, logging.LevelDebug))
			}

			return nil
		},
	}

	cmd.AddCommand(deployCmd.NewCmdDeploy(ctx, f))
	cmd.AddCommand(rollbackCmd.NewCmdRollback(ctx, f))
	cmd.AddCommand(historyCmd.NewCmdHistory(ctx, f))
	cmd.AddCommand(logsCmd.NewCmdLogs(ctx, f))
	cmd.AddCommand(appCmd.NewCmdApp(ctx, f))
	cmd.AddCommand(registryCmd.NewCmdRegistry(ctx, f))
	cmd.AddCommand(proxyCmd.NewCmdProxy(ctx, f))
	cmd.AddCommand(initCmd.NewCmdInit(ctx, f))

	cmd.PersistentFlags().BoolP("debug", "d", false, "Display debugging output in the console")
	viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))
	cmd.PersistentFlags().String("host", "", "Host to run command on")
	viper.BindPFlag("host", cmd.PersistentFlags().Lookup("host"))
	cmd.PersistentFlags().Bool("force", false, "Force non-transactional execution")
	viper.BindPFlag("transaction.bypass", cmd.PersistentFlags().Lookup("force"))

	return cmd
}
