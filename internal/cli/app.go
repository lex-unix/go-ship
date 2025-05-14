package cli

import "github.com/spf13/cobra"

func init() {
	rootCmd.AddCommand(appCmd)
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage application on the servers",
}
