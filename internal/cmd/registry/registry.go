package registry

import "github.com/spf13/cobra"

var RegistryCmd = &cobra.Command{
	Use:   "registry",
	Short: "Login into or logout from registry",
}
