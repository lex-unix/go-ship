package app

import "github.com/spf13/cobra"

var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Manage application on the servers",
}
