package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current version of box, set during build time.
var Version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of box",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println("box", Version)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
