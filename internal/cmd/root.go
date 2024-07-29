package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/cmd/app"
	"neite.dev/go-ship/internal/cmd/registry"
)

func init() {
	rootCmd.AddCommand(app.AppCmd)
	rootCmd.AddCommand(registry.RegistryCmd)
}

var rootCmd = &cobra.Command{
	Use:  "goship",
	Long: "go-ship",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
