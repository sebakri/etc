package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of box",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("box", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
