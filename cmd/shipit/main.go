package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"neite.dev/go-ship/internal/app"
	"neite.dev/go-ship/internal/cli"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/exec/localexec"
	"neite.dev/go-ship/internal/exec/sshexec"
	"neite.dev/go-ship/internal/logging"
	"neite.dev/go-ship/internal/txman"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	if err := config.Load(); err != nil {
		logging.Errorf("failed to load configuration: %s", err)
	}

	cfg := config.Get()
	if cfg.Debug {
		l := logging.New(os.Stderr, logging.LevelDebug)
		logging.SetDefault(l)

	}

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

	f := &cliutil.Factory{
		App:       app,
		Txman:     txmanager,
		Localexec: lexec,
	}

	rootCmd := cli.NewRootCmd(ctx, f)
	if err := rootCmd.Execute(); err != nil {
		logging.Errorf("command failed: %s", err)
		os.Exit(1)
	}
}
