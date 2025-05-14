package cli

import (
	"context"
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
	rootCmd.AddCommand(logCmd)
	logCmd.PersistentFlags().BoolP("follow", "f", false, "Follow logs on servers")
	logCmd.PersistentFlags().IntP("lines", "n", 100, "Number of lines to show from each server")
	logCmd.PersistentFlags().String("since", "", "Show lines since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
}

var logCmd = &cobra.Command{
	Use:   "logs",
	Short: "Fetch logs from you containers",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
		defer cancel()

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
		app := app.New(txmanager, lexec)
		if err := app.Logs(ctx, follow, lines, since); err != nil {
			logging.Errorf("command failed: %s", err)
			os.Exit(1)
		}

		return nil
	},
}
