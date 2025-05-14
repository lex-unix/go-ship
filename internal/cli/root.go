package cli

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"neite.dev/go-ship/internal/config"
	"neite.dev/go-ship/internal/logging"
)

func init() {
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Display debugging output in the console")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().String("host", "", "Host to run command on")
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
		os.Exit(1)
	}
}
