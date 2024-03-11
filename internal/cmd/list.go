package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"neite.dev/go-ship/internal/lockfile"
)

func init() {
	rootCmd.AddCommand(listCmd)
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions of your application",
	Run: func(cmd *cobra.Command, args []string) {
		lockPath, err := lockfile.LockPath()
		if err != nil {
			fmt.Println(err)
			return
		}

		lock, err := os.Open(lockPath)
		if err != nil {
			fmt.Println(err)
			return
		}

		data, err := lockfile.Read(lock)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("Version\t\tImage name\n\n")
		for _, entry := range data {
			fmt.Printf("%s\t\t%s\n", entry.Version, entry.Image)
		}
	},
}
