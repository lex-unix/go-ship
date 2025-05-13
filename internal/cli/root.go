package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/logging"
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Display debugging output in the console. (default: false)")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

var rootCmd = &cobra.Command{
	Use:  "shipit",
	Long: "shipit",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Load(); err != nil {
			return err
		}
		if config.Get().Debug {
			l := logging.New(os.Stderr, logging.LevelDebug)
			logging.SetDefault(l)
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
