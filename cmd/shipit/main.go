package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"neite.dev/go-ship/internal/cli"
	"neite.dev/go-ship/internal/cli/cliutil"
	"neite.dev/go-ship/internal/logging"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	f := cliutil.New()
	rootCmd := cli.NewRootCmd(ctx, f)
	if err := rootCmd.Execute(); err != nil {
		logging.Errorf("command failed: %s", err)
		os.Exit(1)
	}
}
