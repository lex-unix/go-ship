package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lex-unix/faino/internal/cli"
	"github.com/lex-unix/faino/internal/cli/cliutil"
	"github.com/lex-unix/faino/internal/logging"
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
