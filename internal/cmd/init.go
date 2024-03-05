package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/config"
)

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(deployCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if config.IsExists() {
			fmt.Println("config file already exists; skipping")
			return nil
		}

		err := config.NewConfig()
		if err != nil {
			return fmt.Errorf("failed to intialize user config file: %w", err)
		}

		fmt.Println(`Initialized config file "goship.yaml" in the current directory`)

		return nil
	},
}

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Deploy your app on server via Docker",
}
