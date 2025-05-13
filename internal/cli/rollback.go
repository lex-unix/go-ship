package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/txman"
)

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback to your app's desired version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("the version of the app to rollback to is required")
		}

		ctx := context.Background()
		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

		cfg := config.Get()

		txmanager := txman.New()
		for _, server := range cfg.Servers {
			client, err := sshexec.New(server, cfg.SSH.User, cfg.SSH.Port)
			if err != nil {
				logging.Errorf("failed to establish connection with %s: %s", server, err)
				os.Exit(1)
			}
			txmanager.RegisterHost(server, client)
		}

		lexec := localexec.New()
		a := app.New(txmanager, lexec)

		version := args[0]
		if err := a.Rollback(ctx, version); err != nil {
			logging.Errorf("command failed: %s", err)
			os.Exit(1)
		}

		return nil
	},
}
